package middleware

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func Logger(logger *zap.Logger) gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		fields := []zapcore.Field{
			zap.Int("status", param.StatusCode),
			zap.String("method", param.Method),
			zap.String("path", param.Path),
			zap.String("ip", param.ClientIP),
			zap.Duration("latency", param.Latency),
			zap.String("user_agent", param.Request.UserAgent()),
			zap.Time("time", param.TimeStamp),
		}

		if param.ErrorMessage != "" {
			fields = append(fields, zap.String("error", param.ErrorMessage))
		}

		if tenantID := param.Request.Header.Get("X-Tenant-ID"); tenantID != "" {
			fields = append(fields, zap.String("tenant_id", tenantID))
		}

		logger.Info("HTTP Request", fields...)
		return ""
	})
}

func NewLogger(service string, env string) (*zap.Logger, error) {
	config := zap.NewProductionConfig()
	
	if env == "development" {
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	config.InitialFields = map[string]interface{}{
		"service": service,
	}

	return config.Build()
}