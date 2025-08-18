package handlers

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"

	"github.com/zoomxml/internal/services"
)

type CNPJHandler struct {
	cnpjService *services.CNPJService
}

func NewCNPJHandler() *CNPJHandler {
	return &CNPJHandler{
		cnpjService: services.NewCNPJService(),
	}
}

// ConsultarCNPJ consulta dados do CNPJ na API externa
func (h *CNPJHandler) ConsultarCNPJ(c *fiber.Ctx) error {
	cnpj := c.Params("cnpj")
	if cnpj == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "CNPJ é obrigatório",
		})
	}

	log.Info().
		Str("cnpj", cnpj).
		Str("operation", "consultar_cnpj").
		Msg("Iniciando consulta de CNPJ")

	ctx := c.Context()
	cnpjData, err := h.cnpjService.ConsultarCNPJ(ctx, cnpj)
	if err != nil {
		log.Error().
			Err(err).
			Str("cnpj", cnpj).
			Str("operation", "consultar_cnpj").
			Msg("Erro ao consultar CNPJ")

		// Retornar erro específico baseado no tipo
		switch err.Error() {
		case "CNPJ inválido":
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{
				"error": "CNPJ inválido",
			})
		case "CNPJ não encontrado":
			return c.Status(http.StatusNotFound).JSON(fiber.Map{
				"error": "CNPJ não encontrado",
			})
		default:
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"error": "Erro interno do servidor",
			})
		}
	}

	log.Info().
		Str("cnpj", cnpj).
		Str("name", cnpjData.Name).
		Str("operation", "consultar_cnpj").
		Msg("CNPJ consultado com sucesso")

	return c.JSON(cnpjData)
}
