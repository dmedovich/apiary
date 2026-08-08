// Package gin is a minimal stand-in for github.com/gin-gonic/gin. apiary's
// gin-handler detection is purely syntactic (it matches the *gin.Context
// identifier), so this fixture doesn't need the real dependency to exercise
// the annotation-resolution path end to end.
package gin

type Context struct{}

func (c *Context) ShouldBindJSON(obj any) error { return nil }
func (c *Context) JSON(code int, obj any)       {}
