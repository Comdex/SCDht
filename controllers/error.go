package controllers

import (
	"github.com/astaxie/beego"
)

type ErrorController struct {
	beego.Controller
}

func (c *ErrorController) Error404() {
	c.TplNames = "404.html"
}

func (c *ErrorController) Error501() {
	c.TplNames = "501.html"
}
