package controllers

import (
	utils "TaskRestApiService/utils"
	"github.com/gin-gonic/gin"
	"log"
)

type AuthController struct {
	authHost string
}

func NewAuthController(authServiceHost string) *AuthController {
	return &AuthController{authHost: authServiceHost}
}

// User Login
// @Summary      User login
// @Description  Authenticate user with credentials, returns JWT token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        loginRequest  body      map[string]string  true  "Login credentials (e.g. username, password)"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]string
// @Router       /auth/login [post]
func (ac *AuthController) Login(c *gin.Context) {
	log.Printf("received request")
	targetURL := c.DefaultQuery("url", ac.authHost+"/login")

	utils.ProxyRequest(c, targetURL)
}

// User Registration
// @Summary      Register a new user
// @Description  Create a new user account
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        registerRequest  body      map[string]string  true  "User registration data"
// @Success      201  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]string
// @Router       /auth/register [post]
func (ac *AuthController) Register(c *gin.Context) {
	targetURL := c.DefaultQuery("url", ac.authHost+"/register")
	utils.ProxyRequest(c, targetURL)
}

// Validate Token
// @Summary      Validate authentication token
// @Description  Validates the current user's token
// @Tags         auth
// @Produce      json
// @Success      200  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]string
// @Router       /auth/validate [get]
// @Security     BearerAuth
func (ac *AuthController) Validate(c *gin.Context) {
	targetURL := c.DefaultQuery("url", ac.authHost+"/validate")
	utils.ProxyRequest(c, targetURL)
}

// List Users
// @Summary      Get list of users
// @Description  Retrieves list of all users (admin only)
// @Tags         auth
// @Produce      json
// @Success      200  {array}   map[string]interface{}
// @Failure      401  {object}  map[string]string
// @Router       /auth/users [get]
// @Security     BearerAuth
func (ac *AuthController) Users(c *gin.Context) {
	targetURL := c.DefaultQuery("url", ac.authHost+"/users")
	utils.ProxyRequest(c, targetURL)
}

// Get Current User
// @Summary      Get information about current authenticated user
// @Tags         auth
// @Produce      json
// @Success      200  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]string
// @Router       /auth/user [get]
// @Security     BearerAuth
func (ac *AuthController) User(c *gin.Context) {
	targetURL := c.DefaultQuery("url", ac.authHost+"/user")
	utils.ProxyRequest(c, targetURL)
}

// User Logout
// @Summary      Logout current user
// @Description  Invalidate user session/token
// @Tags         auth
// @Produce      json
// @Success      204  "No Content"
// @Router       /auth/logout [post]
// @Security     BearerAuth
func (ac *AuthController) Logout(c *gin.Context) {
	targetURL := c.DefaultQuery("url", ac.authHost+"/logout")
	utils.ProxyRequest(c, targetURL)
}
