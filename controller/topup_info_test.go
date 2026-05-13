package controller

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/setting"
	"github.com/QuantumNous/new-api/setting/operation_setting"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestGetTopUpInfoOmitsEpayMethodsWhenEpayDisabled(t *testing.T) {
	gin.SetMode(gin.TestMode)

	originalPayAddress := operation_setting.PayAddress
	originalEpayID := operation_setting.EpayId
	originalEpayKey := operation_setting.EpayKey
	originalPayMethods := operation_setting.PayMethods
	originalYimaEnabled := setting.YimaEnabled
	originalYimaPayAddress := setting.YimaPayAddress
	originalYimaMerchantID := setting.YimaMerchantId
	originalYimaMerchantKey := setting.YimaMerchantKey
	originalYimaAlipayEnabled := setting.YimaAlipayEnabled
	originalYimaAlipayName := setting.YimaAlipayName
	originalYimaWechatEnabled := setting.YimaWechatEnabled
	originalYimaWechatName := setting.YimaWechatName
	originalYimaMinTopUp := setting.YimaMinTopUp
	t.Cleanup(func() {
		operation_setting.PayAddress = originalPayAddress
		operation_setting.EpayId = originalEpayID
		operation_setting.EpayKey = originalEpayKey
		operation_setting.PayMethods = originalPayMethods
		setting.YimaEnabled = originalYimaEnabled
		setting.YimaPayAddress = originalYimaPayAddress
		setting.YimaMerchantId = originalYimaMerchantID
		setting.YimaMerchantKey = originalYimaMerchantKey
		setting.YimaAlipayEnabled = originalYimaAlipayEnabled
		setting.YimaAlipayName = originalYimaAlipayName
		setting.YimaWechatEnabled = originalYimaWechatEnabled
		setting.YimaWechatName = originalYimaWechatName
		setting.YimaMinTopUp = originalYimaMinTopUp
	})

	operation_setting.PayAddress = ""
	operation_setting.EpayId = ""
	operation_setting.EpayKey = ""
	operation_setting.PayMethods = []map[string]string{{"name": "支付宝", "type": "alipay"}, {"name": "微信", "type": "wxpay"}}

	setting.YimaEnabled = true
	setting.YimaPayAddress = "https://zf.rx.sc.cn"
	setting.YimaMerchantId = "10000001"
	setting.YimaMerchantKey = "secret"
	setting.YimaAlipayEnabled = true
	setting.YimaAlipayName = "支付宝 (倚码)"
	setting.YimaWechatEnabled = true
	setting.YimaWechatName = "微信 (倚码)"
	setting.YimaMinTopUp = 1

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/api/user/topup", nil)

	GetTopUpInfo(ctx)

	require.Equal(t, http.StatusOK, recorder.Code)

	var response struct {
		Success bool `json:"success"`
		Data    struct {
			EnableOnlineTopup bool              `json:"enable_online_topup"`
			EnableYimaTopup   bool              `json:"enable_yima_topup"`
			PayMethods        []map[string]any `json:"pay_methods"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &response))
	require.True(t, response.Success)
	require.True(t, response.Data.EnableOnlineTopup)
	require.True(t, response.Data.EnableYimaTopup)
	require.Len(t, response.Data.PayMethods, 2)
	require.ElementsMatch(t, []string{model.PaymentMethodYimaAlipay, model.PaymentMethodYimaWechat}, []string{
		response.Data.PayMethods[0]["type"].(string),
		response.Data.PayMethods[1]["type"].(string),
	})
	for _, method := range response.Data.PayMethods {
		require.NotEqual(t, "alipay", method["type"])
		require.NotEqual(t, "wxpay", method["type"])
	}
}

func TestGetTopUpInfoUsesConfiguredYimaMethodSettings(t *testing.T) {
	gin.SetMode(gin.TestMode)

	originalYimaEnabled := setting.YimaEnabled
	originalYimaPayAddress := setting.YimaPayAddress
	originalYimaMerchantID := setting.YimaMerchantId
	originalYimaMerchantKey := setting.YimaMerchantKey
	originalYimaAlipayEnabled := setting.YimaAlipayEnabled
	originalYimaAlipayName := setting.YimaAlipayName
	originalYimaWechatEnabled := setting.YimaWechatEnabled
	originalYimaWechatName := setting.YimaWechatName
	originalYimaMinTopUp := setting.YimaMinTopUp
	t.Cleanup(func() {
		setting.YimaEnabled = originalYimaEnabled
		setting.YimaPayAddress = originalYimaPayAddress
		setting.YimaMerchantId = originalYimaMerchantID
		setting.YimaMerchantKey = originalYimaMerchantKey
		setting.YimaAlipayEnabled = originalYimaAlipayEnabled
		setting.YimaAlipayName = originalYimaAlipayName
		setting.YimaWechatEnabled = originalYimaWechatEnabled
		setting.YimaWechatName = originalYimaWechatName
		setting.YimaMinTopUp = originalYimaMinTopUp
	})

	setting.YimaEnabled = true
	setting.YimaPayAddress = "https://zf.rx.sc.cn"
	setting.YimaMerchantId = "10000001"
	setting.YimaMerchantKey = "secret"
	setting.YimaAlipayEnabled = false
	setting.YimaAlipayName = "自定义支付宝"
	setting.YimaWechatEnabled = true
	setting.YimaWechatName = "扫码微信支付"
	setting.YimaMinTopUp = 5

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/api/user/topup", nil)

	GetTopUpInfo(ctx)

	require.Equal(t, http.StatusOK, recorder.Code)

	var response struct {
		Success bool `json:"success"`
		Data    struct {
			EnableYimaTopup bool                `json:"enable_yima_topup"`
			PayMethods      []map[string]string `json:"pay_methods"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &response))
	require.True(t, response.Success)
	require.True(t, response.Data.EnableYimaTopup)
	require.Len(t, response.Data.PayMethods, 1)
	require.Equal(t, model.PaymentMethodYimaWechat, response.Data.PayMethods[0]["type"])
	require.Equal(t, "扫码微信支付", response.Data.PayMethods[0]["name"])
	require.Equal(t, "5", response.Data.PayMethods[0]["min_topup"])
}