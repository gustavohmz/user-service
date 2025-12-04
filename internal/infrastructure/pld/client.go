package pld

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"user-service/internal/domain"
	"go.uber.org/zap"
)

type pldClient struct {
	baseURL string
	timeout time.Duration
	client  *http.Client
	logger  *zap.Logger
}

func NewPLDClient(baseURL string, timeoutSeconds int, logger *zap.Logger) domain.PLDService {
	return &pldClient{
		baseURL: baseURL,
		timeout: time.Duration(timeoutSeconds) * time.Second,
		client: &http.Client{
			Timeout: time.Duration(timeoutSeconds) * time.Second,
		},
		logger: logger,
	}
}

type PLDRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
}

type PLDResponse struct {
	IsInBlacklist bool `json:"is_in_blacklist"`
}

func (c *pldClient) CheckBlacklist(ctx context.Context, firstName, lastName, email string) (bool, error) {
	url := fmt.Sprintf("%s/check-blacklist", c.baseURL)

	requestBody := PLDRequest{
		FirstName: firstName,
		LastName:  lastName,
		Email:      email,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		if c.logger != nil {
			c.logger.Error("Error al serializar request PLD",
				zap.String("email", email),
				zap.Error(err),
			)
		}
		return false, fmt.Errorf("error al serializar request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		if c.logger != nil {
			c.logger.Error("Error al crear request PLD",
				zap.String("email", email),
				zap.String("url", url),
				zap.Error(err),
			)
		}
		return false, fmt.Errorf("error al crear request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	if c.logger != nil {
		c.logger.Info("Consultando servicio PLD",
			zap.String("email", email),
			zap.String("first_name", firstName),
			zap.String("last_name", lastName),
		)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		if c.logger != nil {
			c.logger.Warn("Error al consultar servicio PLD",
				zap.String("email", email),
				zap.Error(err),
			)
		}
		return false, nil
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		if c.logger != nil {
			c.logger.Warn("Error al leer respuesta PLD",
				zap.String("email", email),
				zap.Int("status_code", resp.StatusCode),
				zap.Error(err),
			)
		}
		return false, nil
	}

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		if c.logger != nil {
			c.logger.Warn("Servicio PLD retornó error",
				zap.String("email", email),
				zap.Int("status_code", resp.StatusCode),
				zap.String("response_body", string(body)),
			)
		}
		return false, nil
	}

	var pldResp PLDResponse
	if err := json.Unmarshal(body, &pldResp); err != nil {
		if c.logger != nil {
			c.logger.Warn("Error al parsear respuesta PLD",
				zap.String("email", email),
				zap.String("response_body", string(body)),
				zap.Error(err),
			)
		}
		return false, nil
	}

	if c.logger != nil {
		c.logger.Info("Verificación PLD completada",
			zap.String("email", email),
			zap.Bool("is_in_blacklist", pldResp.IsInBlacklist),
		)
	}

	return pldResp.IsInBlacklist, nil
}

