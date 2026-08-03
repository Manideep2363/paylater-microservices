package response

import "github.com/gin-gonic/gin"

// Error writes a standard JSON error payload: {"error": "..."}.
func Error(c *gin.Context, status int, message string) {
	c.JSON(status, gin.H{"error": message})
}

// Message writes a standard JSON message payload: {"message": "..."}.
func Message(c *gin.Context, status int, message string) {
	c.JSON(status, gin.H{"message": message})
}

// JSON writes an arbitrary JSON payload with the given status code.
func JSON(c *gin.Context, status int, data any) {
	c.JSON(status, data)
}
