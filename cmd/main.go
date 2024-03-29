package main

import (
	"html/template"
	"io"
	"slices"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type Template struct {
	templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, ctx echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func newTemplate() *Template {
	return &Template{
		templates: template.Must(template.ParseGlob("views/*.html")),
	}
}

var id = 0

type Contact struct {
	Name  string
	Email string
	Id    int
}

func newContact(name, email string) Contact {
	id++
	return Contact{
		Name:  name,
		Email: email,
		Id:    id,
	}
}

type Contacts = []Contact

type DB struct {
	Contacts Contacts
}

func (d *DB) hasEmail(email string) bool {
	for _, contact := range d.Contacts {
		if contact.Email == email {
			return true
		}
	}
	return false
}

func (d *DB) indexOf(id int) int {
	for i, contact := range d.Contacts {
		if contact.Id == id {
			return i
		}
	}
	return -1
}

func newDB() DB {
	return DB{
		Contacts: []Contact{
			newContact("John Doe", "jdoe@gmail.com"),
			newContact("Clara Doe", "cd@gmail.com"),
		},
	}
}

type FormData struct {
	Values map[string]string
	Errors map[string]string
}

func newFormData() FormData {
	return FormData{
		Values: make(map[string]string),
		Errors: make(map[string]string),
	}
}

type Page struct {
	DB   DB
	Form FormData
}

func newPage() Page {
	return Page{
		DB:   newDB(),
		Form: newFormData(),
	}
}

func main() {
	e := echo.New()
	e.Use(middleware.Logger())

	page := newPage()
	e.Renderer = newTemplate()

	e.GET("/", func(c echo.Context) error {
		return c.Render(200, "index", page)
	})

	e.POST("/contacts", func(c echo.Context) error {
		name := c.FormValue("name")
		email := c.FormValue("email")

		if page.DB.hasEmail(email) {
			formData := newFormData()
			formData.Values["name"] = name
			formData.Values["email"] = email

			formData.Errors["email"] = "Email already Exists"

			// NOTE: fckin retard, HTMX not rendering statuscode = 400's
			// find out later why and fix it. 200 is not good but it works
			return c.Render(200, "form", formData)
		}

		contact := newContact(name, email)
		page.DB.Contacts = append(page.DB.Contacts, contact)

		c.Render(200, "form", newFormData())
		return c.Render(200, "oob-contact", contact)
	})

	e.DELETE("/contacts/:id", func(c echo.Context) error {
		idStr := c.Param("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			return c.String(400, "Invalid Id")
		}

		index := page.DB.indexOf(id)
		if index == -1 {
			return c.String(400, "Contact does not exist")
		}

		page.DB.Contacts = slices.Delete(page.DB.Contacts, index, index+1)
		return c.NoContent(200)
	})

	e.Logger.Fatal(e.Start(":42069"))
}
