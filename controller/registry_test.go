package controller

import (
	"testing"

	"github.com/kataras/iris/v12"
)

type TestController struct {
	Base
}

func (c *TestController) GetConfiguration() Configuration {
	return Configuration{
		Name: "Test",
	}
}

func (c *TestController) Index(ctx iris.Context) {
	ctx.WriteString("index")
}

func (c *TestController) Show(ctx iris.Context) {
	ctx.WriteString("show")
}

func (c *TestController) NoContext() {
	// This method should not be included
}

func TestRegistry_Register(t *testing.T) {
	r := NewRegistry()
	if r == nil {
		t.Fatal("expected non-nil registry")
	}

	r.Register(&TestController{})

	handlers := r.Handlers()
	if len(handlers) != 2 {
		t.Errorf("expected 2 handlers, got %d", len(handlers))
	}

	// Check handler names
	handlerNames := make(map[string]bool)
	for _, h := range handlers {
		handlerNames[h.Method] = true
	}

	if !handlerNames["Index"] {
		t.Error("expected Index handler")
	}
	if !handlerNames["Show"] {
		t.Error("expected Show handler")
	}
	if handlerNames["NoContext"] {
		t.Error("should not include NoContext handler")
	}
}

func TestRegistry_Handlers_WithName(t *testing.T) {
	r := NewRegistry()
	r.Register(&TestController{})

	handlers := r.Handlers()

	for _, h := range handlers {
		if h.Controller != "Test" {
			t.Errorf("expected controller name 'Test', got %s", h.Controller)
		}
	}
}

func TestRegistry_Handlers_HandlerType(t *testing.T) {
	r := NewRegistry()
	r.Register(&TestController{})

	handlers := r.Handlers()

	for _, h := range handlers {
		if h.Handler == nil {
			t.Error("expected non-nil handler")
		}
	}
}
