package model

import (
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/setting"
	"github.com/QuantumNous/new-api/setting/operation_setting"
	"github.com/stretchr/testify/require"
)

func TestUpdateOptionMapNormalizesNilString(t *testing.T) {
	originalOptionMap := common.OptionMap
	originalEpayID := operation_setting.EpayId
	t.Cleanup(func() {
		common.OptionMap = originalOptionMap
		operation_setting.EpayId = originalEpayID
	})

	common.OptionMap = map[string]string{}
	operation_setting.EpayId = "old-value"

	require.NoError(t, updateOptionMap("EpayId", "<nil>"))
	require.Equal(t, "", common.OptionMap["EpayId"])
	require.Equal(t, "", operation_setting.EpayId)

	require.NoError(t, updateOptionMap("EpayId", "actual-id"))
	require.Equal(t, "actual-id", common.OptionMap["EpayId"])
	require.Equal(t, "actual-id", operation_setting.EpayId)
}

func TestUpdateOptionMapUpdatesYimaMethodSettings(t *testing.T) {
	originalOptionMap := common.OptionMap
	originalYimaAlipayEnabled := setting.YimaAlipayEnabled
	originalYimaAlipayName := setting.YimaAlipayName
	originalYimaWechatEnabled := setting.YimaWechatEnabled
	originalYimaWechatName := setting.YimaWechatName
	t.Cleanup(func() {
		common.OptionMap = originalOptionMap
		setting.YimaAlipayEnabled = originalYimaAlipayEnabled
		setting.YimaAlipayName = originalYimaAlipayName
		setting.YimaWechatEnabled = originalYimaWechatEnabled
		setting.YimaWechatName = originalYimaWechatName
	})

	common.OptionMap = map[string]string{}
	setting.YimaAlipayEnabled = true
	setting.YimaAlipayName = "支付宝 (倚码)"
	setting.YimaWechatEnabled = true
	setting.YimaWechatName = "微信 (倚码)"

	require.NoError(t, updateOptionMap("YimaAlipayEnabled", "false"))
	require.NoError(t, updateOptionMap("YimaAlipayName", "自定义支付宝"))
	require.NoError(t, updateOptionMap("YimaWechatEnabled", "true"))
	require.NoError(t, updateOptionMap("YimaWechatName", "扫码微信支付"))

	require.Equal(t, "false", common.OptionMap["YimaAlipayEnabled"])
	require.False(t, setting.YimaAlipayEnabled)
	require.Equal(t, "自定义支付宝", common.OptionMap["YimaAlipayName"])
	require.Equal(t, "自定义支付宝", setting.YimaAlipayName)
	require.Equal(t, "true", common.OptionMap["YimaWechatEnabled"])
	require.True(t, setting.YimaWechatEnabled)
	require.Equal(t, "扫码微信支付", common.OptionMap["YimaWechatName"])
	require.Equal(t, "扫码微信支付", setting.YimaWechatName)
}