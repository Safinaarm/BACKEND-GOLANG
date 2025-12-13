package service

import "github.com/gofiber/fiber/v2"

type UserRouteService interface {
	ListUsersHandler(c *fiber.Ctx) error
	GetUserHandler(c *fiber.Ctx) error
	CreateUserHandler(c *fiber.Ctx) error
	UpdateUserHandler(c *fiber.Ctx) error
	DeleteUserHandler(c *fiber.Ctx) error
	UpdateUserRoleHandler(c *fiber.Ctx) error
}
