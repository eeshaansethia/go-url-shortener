package routes

import (
	"go-url-shortener/database"
	"go-url-shortener/helpers"
	"os"
	"strconv"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"
)

type request struct {
	URL         string 			`json:"url"`
	CustomShort string 			`json:"short"`
	Expiry      time.Duration  	`json:"expiry"`
}

type response struct {
	URL            string 			`json:"url"`
	CustomShort    string 			`json:"short"`
	Expiry         time.Duration  	`json:"expiry"`
	XRateRemaining int  			`json:"rate_limit"`
	XRateLimitRest time.Duration	`json:"rate_limit_reset"`
}

func ShortenURL(c *fiber.Ctx) error{
	body := new(request)
	if err := c.BodyParser(body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request",
		})
	}

	// Rate Limiting
	r2 := database.CreateClient(1)
	defer r2.Close()
	val, err := r2.Get(database.Ctx, c.IP()).Result()
	if err == redis.Nil {
		err = r2.Set(database.Ctx, c.IP(), os.Getenv("API_QOUTA"), 30*60*time.Second).Err()
	} else {
		valInt, _ := strconv.Atoi(val)
		if valInt <= 0 {
			limit, _ := r2.TTL(database.Ctx, c.IP()).Result()
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error": "Rate limit exceeded",
				"rate_limit": limit,
				"rate_limit_reset": limit/time.Nanosecond/time.Minute,
			})
		}
	}

	if !govalidator.IsURL(body.URL){
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid URL",
		})
	}

	if !helpers.RemoveDomainError(body.URL){
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid URL",
		})
	}

	body.URL = helpers.EnforceHTTP(body.URL)
}