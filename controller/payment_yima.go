package controller

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/logger"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/service"
	"github.com/QuantumNous/new-api/setting"
	"github.com/QuantumNous/new-api/setting/operation_setting"
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"github.com/samber/lo"
)

const (
	yimaStatusPaid    = "paid"
	yimaStatusAwait   = "awaiting"
	yimaStatusCreated = "created"
)

type yimaCreateResponse struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    yimaOrderDetail `json:"data"`
}

type yimaQueryResponse struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    yimaOrderDetail `json:"data"`
}

type yimaOrderDetail struct {
	OrderNo     string `json:"order_no"`
	OutTradeNo  string `json:"out_trade_no"`
	PayURL      string `json:"pay_url"`
	Amount      string `json:"amount"`
	AmountPay   string `json:"amount_pay"`
	PayMethod   string `json:"pay_method"`
	Status      string `json:"status"`
	ExpireAt    string `json:"expire_at"`
	PaidAt      string `json:"paid_at"`
	Description string `json:"description"`
	Title       string `json:"title"`
}

func isYimaPaymentMethod(method string) bool {
	return method == model.PaymentMethodYimaAlipay || method == model.PaymentMethodYimaWechat
}

func yimaDefaultMethodName(method string) string {
	switch method {
	case model.PaymentMethodYimaAlipay:
		return "支付宝 (倚码)"
	case model.PaymentMethodYimaWechat:
		return "微信 (倚码)"
	default:
		return ""
	}
}

func getYimaMethodName(method string) string {
	var configured string
	switch method {
	case model.PaymentMethodYimaAlipay:
		configured = setting.YimaAlipayName
	case model.PaymentMethodYimaWechat:
		configured = setting.YimaWechatName
	}
	configured = strings.TrimSpace(configured)
	if configured != "" {
		return configured
	}
	return yimaDefaultMethodName(method)
}

func isYimaMethodEnabled(method string) bool {
	switch method {
	case model.PaymentMethodYimaAlipay:
		return setting.YimaAlipayEnabled
	case model.PaymentMethodYimaWechat:
		return setting.YimaWechatEnabled
	default:
		return false
	}
}

func getEnabledYimaMethodTypes() []string {
	methodTypes := []string{model.PaymentMethodYimaAlipay, model.PaymentMethodYimaWechat}
	enabledMethods := make([]string, 0, len(methodTypes))
	for _, methodType := range methodTypes {
		if isYimaMethodEnabled(methodType) {
			enabledMethods = append(enabledMethods, methodType)
		}
	}
	return enabledMethods
}

func yimaGatewayMethod(method string) string {
	switch method {
	case model.PaymentMethodYimaAlipay:
		return "alipay"
	case model.PaymentMethodYimaWechat:
		return "wechat"
	default:
		return ""
	}
}

func normalizeYimaBaseURL() string {
	return strings.TrimRight(strings.TrimSpace(setting.YimaPayAddress), "/")
}

func getYimaTopUpNotifyURL() string {
	if strings.TrimSpace(setting.YimaNotifyUrl) != "" {
		return strings.TrimSpace(setting.YimaNotifyUrl)
	}
	return strings.TrimRight(service.GetCallbackAddress(), "/") + "/api/user/yima/notify"
}

func getYimaTopUpReturnURL() string {
	if strings.TrimSpace(setting.YimaReturnUrl) != "" {
		return strings.TrimSpace(setting.YimaReturnUrl)
	}
	return paymentReturnPath("/api/user/yima/return")
}

func getYimaSubscriptionReturnURL() string {
	if strings.TrimSpace(setting.YimaSubscriptionReturnUrl) != "" {
		return strings.TrimSpace(setting.YimaSubscriptionReturnUrl)
	}
	if strings.TrimSpace(setting.YimaReturnUrl) != "" {
		return strings.TrimSpace(setting.YimaReturnUrl)
	}
	return paymentReturnPath("/api/subscription/yima/return")
}

func parseYimaParams(c *gin.Context) (map[string]string, error) {
	if c.Request.Method == http.MethodPost {
		if err := c.Request.ParseForm(); err != nil {
			return nil, err
		}
		return lo.Reduce(lo.Keys(c.Request.PostForm), func(result map[string]string, key string, _ int) map[string]string {
			result[key] = c.Request.PostForm.Get(key)
			return result
		}, map[string]string{}), nil
	}

	return lo.Reduce(lo.Keys(c.Request.URL.Query()), func(result map[string]string, key string, _ int) map[string]string {
		result[key] = c.Request.URL.Query().Get(key)
		return result
	}, map[string]string{}), nil
}

func buildYimaSignature(params map[string]string, mchKey string) string {
	keys := make([]string, 0, len(params))
	for key := range params {
		if key == "signature" {
			continue
		}
		keys = append(keys, key)
	}
	sort.Strings(keys)
	pairs := make([]string, 0, len(keys))
	for _, key := range keys {
		pairs = append(pairs, key+"="+params[key])
	}
	raw := strings.Join(pairs, "&") + mchKey
	sum := md5.Sum([]byte(raw))
	return hex.EncodeToString(sum[:])
}

func verifyYimaSignature(params map[string]string) bool {
	provided := strings.TrimSpace(params["signature"])
	if provided == "" {
		return false
	}
	return strings.EqualFold(provided, buildYimaSignature(params, setting.YimaMerchantKey))
}

func absoluteYimaPayURL(path string) string {
	trimmed := strings.TrimSpace(path)
	if trimmed == "" {
		return ""
	}
	if strings.HasPrefix(trimmed, "http://") || strings.HasPrefix(trimmed, "https://") {
		return trimmed
	}
	return normalizeYimaBaseURL() + trimmed
}

func submitYimaCreateOrder(form map[string]string) (*yimaOrderDetail, string, error) {
	form["signature"] = buildYimaSignature(form, setting.YimaMerchantKey)
	body := url.Values{}
	for key, value := range form {
		body.Set(key, value)
	}
	req, err := http.NewRequest(http.MethodPost, normalizeYimaBaseURL()+"/v1/merchant-openapi/pay/create", strings.NewReader(body.Encode()))
	if err != nil {
		return nil, "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := (&http.Client{Timeout: 15 * time.Second}).Do(req)
	if err != nil {
		return nil, "", err
	}
	defer service.CloseResponseBodyGracefully(resp)

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", err
	}

	var result yimaCreateResponse
	if err := common.Unmarshal(raw, &result); err != nil {
		return nil, string(raw), err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, string(raw), fmt.Errorf("Yima create order http status %d", resp.StatusCode)
	}
	if result.Code != 0 {
		return nil, string(raw), fmt.Errorf("Yima create order failed: %s", result.Message)
	}
	return &result.Data, string(raw), nil
}

func queryYimaOrder(outTradeNo string) (*yimaOrderDetail, string, error) {
	params := map[string]string{
		"mch_id":       setting.YimaMerchantId,
		"out_trade_no": outTradeNo,
	}
	params["signature"] = buildYimaSignature(params, setting.YimaMerchantKey)
	query := url.Values{}
	for key, value := range params {
		query.Set(key, value)
	}
	endpoint := normalizeYimaBaseURL() + "/v1/merchant-openapi/order/query?" + query.Encode()
	resp, err := (&http.Client{Timeout: 15 * time.Second}).Get(endpoint)
	if err != nil {
		return nil, "", err
	}
	defer service.CloseResponseBodyGracefully(resp)

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", err
	}

	var result yimaQueryResponse
	if err := common.Unmarshal(raw, &result); err != nil {
		return nil, string(raw), err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, string(raw), fmt.Errorf("Yima query order http status %d", resp.StatusCode)
	}
	if result.Code != 0 {
		return nil, string(raw), fmt.Errorf("Yima query order failed: %s", result.Message)
	}
	return &result.Data, string(raw), nil
}

func getYimaPaidAmount(params map[string]string) string {
	if amountPay := strings.TrimSpace(params["amount_pay"]); amountPay != "" {
		return amountPay
	}
	return strings.TrimSpace(params["amount"])
}

func yimaAmountsMatch(expected float64, actual string) bool {
	if strings.TrimSpace(actual) == "" {
		return false
	}
	expectedDecimal := decimal.NewFromFloat(expected).Round(2)
	actualDecimal, err := decimal.NewFromString(actual)
	if err != nil {
		return false
	}
	return expectedDecimal.Equal(actualDecimal.Round(2))
}

func requestYimaTopUp(c *gin.Context, userID int, amount int64, paymentMethod string, payMoney float64) {
	if !isYimaTopUpEnabled() {
		c.JSON(http.StatusOK, gin.H{"message": "error", "data": "倚码支付未启用"})
		return
	}
	if amount < int64(setting.YimaMinTopUp) {
		c.JSON(http.StatusOK, gin.H{"message": "error", "data": fmt.Sprintf("充值数量不能小于 %d", setting.YimaMinTopUp)})
		return
	}

	if !isYimaPaymentMethod(paymentMethod) {
		c.JSON(http.StatusOK, gin.H{"message": "error", "data": "不支持的倚码支付方式"})
		return
	}
	if !isYimaMethodEnabled(paymentMethod) {
		c.JSON(http.StatusOK, gin.H{"message": "error", "data": "该倚码支付方式未启用"})
		return
	}

	providerMethod := yimaGatewayMethod(paymentMethod)

	tradeNo := fmt.Sprintf("USR%dYM%d", userID, time.Now().UnixNano())
	normalizedAmount := amount
	if operation_setting.GetQuotaDisplayType() == operation_setting.QuotaDisplayTypeTokens {
		dAmount := decimal.NewFromInt(amount)
		dQuotaPerUnit := decimal.NewFromFloat(common.QuotaPerUnit)
		normalizedAmount = dAmount.Div(dQuotaPerUnit).IntPart()
	}
	topUp := &model.TopUp{
		UserId:          userID,
		Amount:          normalizedAmount,
		Money:           payMoney,
		TradeNo:         tradeNo,
		PaymentMethod:   providerMethod,
		PaymentProvider: model.PaymentProviderYima,
		CreateTime:      time.Now().Unix(),
		Status:          common.TopUpStatusPending,
	}
	if err := topUp.Insert(); err != nil {
		logger.LogError(c.Request.Context(), fmt.Sprintf("倚码支付 创建充值订单失败 user_id=%d trade_no=%s amount=%d error=%q", userID, tradeNo, amount, err.Error()))
		c.JSON(http.StatusOK, gin.H{"message": "error", "data": "创建订单失败"})
		return
	}

	requestForm := map[string]string{
		"mch_id":       setting.YimaMerchantId,
		"pay_method":   providerMethod,
		"out_trade_no": tradeNo,
		"notify_url":   getYimaTopUpNotifyURL(),
		"return_url":   getYimaTopUpReturnURL(),
		"title":        fmt.Sprintf("TUC%d", amount),
		"description":  fmt.Sprintf("Top up %d", amount),
		"amount":       decimal.NewFromFloat(payMoney).Round(2).StringFixed(2),
	}
	orderData, rawResponse, err := submitYimaCreateOrder(requestForm)
	if err != nil {
		topUp.Status = common.TopUpStatusFailed
		_ = topUp.Update()
		logger.LogError(c.Request.Context(), fmt.Sprintf("倚码支付 创建上游订单失败 user_id=%d trade_no=%s amount=%d error=%q response=%q", userID, tradeNo, amount, err.Error(), rawResponse))
		c.JSON(http.StatusOK, gin.H{"message": "error", "data": "拉起支付失败"})
		return
	}

	paymentURL := absoluteYimaPayURL(orderData.PayURL)
	logger.LogInfo(c.Request.Context(), fmt.Sprintf("倚码支付 充值订单创建成功 user_id=%d trade_no=%s amount=%d money=%.2f payment_method=%s response=%q", userID, tradeNo, amount, payMoney, providerMethod, rawResponse))
	c.JSON(http.StatusOK, gin.H{"message": "success", "data": gin.H{"payment_url": paymentURL}})
}

func completeYimaTopUp(c *gin.Context, params map[string]string) error {
	tradeNo := strings.TrimSpace(params["out_trade_no"])
	if tradeNo == "" {
		return fmt.Errorf("missing out_trade_no")
	}
	LockOrder(tradeNo)
	defer UnlockOrder(tradeNo)

	topUp := model.GetTopUpByTradeNo(tradeNo)
	if topUp == nil {
		return model.ErrTopUpNotFound
	}
	if topUp.PaymentProvider != model.PaymentProviderYima {
		return model.ErrPaymentMethodMismatch
	}
	if !yimaAmountsMatch(topUp.Money, getYimaPaidAmount(params)) {
		return fmt.Errorf("amount mismatch")
	}
	if topUp.Status != common.TopUpStatusPending {
		return nil
	}

	topUp.Status = common.TopUpStatusSuccess
	topUp.CompleteTime = common.GetTimestamp()
	callbackMethod := strings.TrimSpace(params["pay_method"])
	if callbackMethod != "" && topUp.PaymentMethod != callbackMethod {
		topUp.PaymentMethod = callbackMethod
	}
	if err := topUp.Update(); err != nil {
		return err
	}

	dAmount := decimal.NewFromInt(topUp.Amount)
	dQuotaPerUnit := decimal.NewFromFloat(common.QuotaPerUnit)
	quotaToAdd := int(dAmount.Mul(dQuotaPerUnit).IntPart())
	if err := model.IncreaseUserQuota(topUp.UserId, quotaToAdd, true); err != nil {
		return err
	}
	model.RecordTopupLog(topUp.UserId, fmt.Sprintf("使用在线充值成功，充值金额: %v，支付金额：%f", logger.LogQuota(quotaToAdd), topUp.Money), c.ClientIP(), topUp.PaymentMethod, callbackMethod)
	logger.LogInfo(c.Request.Context(), fmt.Sprintf("倚码支付 充值成功 trade_no=%s user_id=%d client_ip=%s quota_to_add=%d money=%.2f", topUp.TradeNo, topUp.UserId, c.ClientIP(), quotaToAdd, topUp.Money))
	return nil
}

func YimaWebhook(c *gin.Context) {
	YimaTopUpNotify(c)
}

func YimaTopUpNotify(c *gin.Context) {
	if !isYimaWebhookEnabled() {
		logger.LogWarn(c.Request.Context(), fmt.Sprintf("倚码支付 webhook 被拒绝 reason=webhook_disabled path=%q client_ip=%s", c.Request.RequestURI, c.ClientIP()))
		_, _ = c.Writer.Write([]byte("fail"))
		return
	}
	params, err := parseYimaParams(c)
	if err != nil {
		logger.LogError(c.Request.Context(), fmt.Sprintf("倚码支付 webhook 参数解析失败 path=%q client_ip=%s error=%q", c.Request.RequestURI, c.ClientIP(), err.Error()))
		_, _ = c.Writer.Write([]byte("fail"))
		return
	}
	logger.LogInfo(c.Request.Context(), fmt.Sprintf("倚码支付 webhook 收到请求 path=%q client_ip=%s params=%q", c.Request.RequestURI, c.ClientIP(), common.GetJsonString(params)))
	if len(params) == 0 || !verifyYimaSignature(params) {
		logger.LogWarn(c.Request.Context(), fmt.Sprintf("倚码支付 webhook 验签失败 path=%q client_ip=%s", c.Request.RequestURI, c.ClientIP()))
		_, _ = c.Writer.Write([]byte("fail"))
		return
	}
	if strings.TrimSpace(params["status"]) != yimaStatusPaid {
		logger.LogInfo(c.Request.Context(), fmt.Sprintf("倚码支付 webhook 忽略事件 trade_no=%s status=%s client_ip=%s", params["out_trade_no"], params["status"], c.ClientIP()))
		_, _ = c.Writer.Write([]byte("success"))
		return
	}
	if err := completeYimaTopUp(c, params); err != nil {
		logger.LogError(c.Request.Context(), fmt.Sprintf("倚码支付 充值处理失败 trade_no=%s client_ip=%s error=%q", params["out_trade_no"], c.ClientIP(), err.Error()))
		_, _ = c.Writer.Write([]byte("fail"))
		return
	}
	_, _ = c.Writer.Write([]byte("success"))
}

func YimaTopUpReturn(c *gin.Context) {
	params, err := parseYimaParams(c)
	if err != nil || len(params) == 0 || !verifyYimaSignature(params) {
		c.Redirect(http.StatusFound, paymentReturnPath("/console/topup?pay=fail"))
		return
	}
	if strings.TrimSpace(params["status"]) == yimaStatusPaid {
		if err := completeYimaTopUp(c, params); err != nil && err != model.ErrTopUpStatusInvalid {
			logger.LogWarn(c.Request.Context(), fmt.Sprintf("倚码支付 return 完成充值失败 trade_no=%s error=%q", params["out_trade_no"], err.Error()))
			c.Redirect(http.StatusFound, paymentReturnPath("/console/topup?pay=fail"))
			return
		}
		c.Redirect(http.StatusFound, paymentReturnPath("/console/topup?pay=success"))
		return
	}
	c.Redirect(http.StatusFound, paymentReturnPath("/console/topup?pay=pending"))
}

func subscriptionRequestYima(c *gin.Context, planID int, paymentMethod string, plan *model.SubscriptionPlan) {
	if !isYimaTopUpEnabled() {
		common.ApiErrorMsg(c, "倚码支付未启用")
		return
	}
	if plan == nil {
		common.ApiErrorMsg(c, "套餐不存在")
		return
	}
	if !isYimaPaymentMethod(paymentMethod) {
		common.ApiErrorMsg(c, "不支持的倚码支付方式")
		return
	}
	if !isYimaMethodEnabled(paymentMethod) {
		common.ApiErrorMsg(c, "该倚码支付方式未启用")
		return
	}
	providerMethod := yimaGatewayMethod(paymentMethod)
	userID := c.GetInt("id")
	tradeNo := fmt.Sprintf("SUBUSR%dYM%d", userID, time.Now().UnixNano())
	order := &model.SubscriptionOrder{
		UserId:          userID,
		PlanId:          planID,
		Money:           plan.PriceAmount,
		TradeNo:         tradeNo,
		PaymentMethod:   providerMethod,
		PaymentProvider: model.PaymentProviderYima,
		CreateTime:      time.Now().Unix(),
		Status:          common.TopUpStatusPending,
	}
	if err := order.Insert(); err != nil {
		common.ApiErrorMsg(c, "创建订单失败")
		return
	}
	requestForm := map[string]string{
		"mch_id":       setting.YimaMerchantId,
		"pay_method":   providerMethod,
		"out_trade_no": tradeNo,
		"notify_url":   strings.TrimRight(service.GetCallbackAddress(), "/") + "/api/subscription/yima/notify",
		"return_url":   getYimaSubscriptionReturnURL(),
		"title":        plan.Title,
		"description":  plan.Title,
		"amount":       decimal.NewFromFloat(plan.PriceAmount).Round(2).StringFixed(2),
	}
	orderData, rawResponse, err := submitYimaCreateOrder(requestForm)
	if err != nil {
		_ = model.ExpireSubscriptionOrder(tradeNo, model.PaymentProviderYima)
		logger.LogError(c.Request.Context(), fmt.Sprintf("倚码支付 创建订阅订单失败 user_id=%d trade_no=%s plan_id=%d error=%q response=%q", userID, tradeNo, planID, err.Error(), rawResponse))
		common.ApiErrorMsg(c, "拉起支付失败")
		return
	}
	logger.LogInfo(c.Request.Context(), fmt.Sprintf("倚码支付 订阅订单创建成功 user_id=%d trade_no=%s plan_id=%d payment_method=%s response=%q", userID, tradeNo, planID, providerMethod, rawResponse))
	c.JSON(http.StatusOK, gin.H{"message": "success", "data": gin.H{"payment_url": absoluteYimaPayURL(orderData.PayURL)}})
}

func completeYimaSubscription(c *gin.Context, params map[string]string) error {
	tradeNo := strings.TrimSpace(params["out_trade_no"])
	if tradeNo == "" {
		return fmt.Errorf("missing out_trade_no")
	}
	LockOrder(tradeNo)
	defer UnlockOrder(tradeNo)
	payload, _ := common.Marshal(params)
	return model.CompleteSubscriptionOrder(tradeNo, string(payload), model.PaymentProviderYima, strings.TrimSpace(params["pay_method"]))
}

func YimaSubscriptionNotify(c *gin.Context) {
	params, err := parseYimaParams(c)
	if err != nil || len(params) == 0 || !verifyYimaSignature(params) {
		_, _ = c.Writer.Write([]byte("fail"))
		return
	}
	if strings.TrimSpace(params["status"]) != yimaStatusPaid {
		_, _ = c.Writer.Write([]byte("success"))
		return
	}
	if orderData, rawResponse, err := queryYimaOrder(strings.TrimSpace(params["out_trade_no"])); err == nil {
		if orderData != nil && orderData.Status == yimaStatusPaid {
			params["pay_method"] = strings.TrimSpace(orderData.PayMethod)
			params["amount_pay"] = strings.TrimSpace(orderData.AmountPay)
			logger.LogInfo(c.Request.Context(), fmt.Sprintf("倚码支付 订阅订单查单确认 trade_no=%s response=%q", params["out_trade_no"], rawResponse))
		}
	}
	if err := completeYimaSubscription(c, params); err != nil {
		logger.LogError(c.Request.Context(), fmt.Sprintf("倚码支付 订阅支付处理失败 trade_no=%s client_ip=%s error=%q", params["out_trade_no"], c.ClientIP(), err.Error()))
		_, _ = c.Writer.Write([]byte("fail"))
		return
	}
	_, _ = c.Writer.Write([]byte("success"))
}

func YimaSubscriptionReturn(c *gin.Context) {
	params, err := parseYimaParams(c)
	if err != nil || len(params) == 0 || !verifyYimaSignature(params) {
		c.Redirect(http.StatusFound, paymentReturnPath("/console/topup?pay=fail"))
		return
	}
	if strings.TrimSpace(params["status"]) == yimaStatusPaid {
		if err := completeYimaSubscription(c, params); err != nil {
			c.Redirect(http.StatusFound, paymentReturnPath("/console/topup?pay=fail"))
			return
		}
		c.Redirect(http.StatusFound, paymentReturnPath("/console/topup?pay=success"))
		return
	}
	if strings.TrimSpace(params["status"]) == yimaStatusAwait || strings.TrimSpace(params["status"]) == yimaStatusCreated {
		c.Redirect(http.StatusFound, paymentReturnPath("/console/topup?pay=pending"))
		return
	}
	c.Redirect(http.StatusFound, paymentReturnPath("/console/topup?pay=fail"))
}
