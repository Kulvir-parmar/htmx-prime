package main

import (
	"html/template"
	"io"

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

type Contact struct {
	Name  string
	Email string
}

func newContact(name, email string) Contact {
	return Contact{
		Name:  name,
		Email: email,
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

func main() {
	e := echo.New()
	e.Use(middleware.Logger())

	myDB := newDB()
	e.Renderer = newTemplate()

	e.GET("/", func(c echo.Context) error {
		return c.Render(200, "index", myDB)
	})

	e.POST("/contacts", func(c echo.Context) error {
		name := c.FormValue("name")
		email := c.FormValue("email")

		if myDB.hasEmail(email) {
			formData := newFormData()
			formData.Values["email"] = name
			formData.Values["email"] = email
			formData.Errors["email"] = "Email already Exists"

			return c.Render(400, "form", formData)
		}

		myDB.Contacts = append(myDB.Contacts, newContact(name, email))
		return c.Render(200, "contacts", myDB)
	})

	e.Logger.Fatal(e.Start(":42069"))
}
