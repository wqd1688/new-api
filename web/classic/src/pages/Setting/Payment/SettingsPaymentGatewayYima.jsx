/*
Copyright (C) 2025 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.

For commercial licensing, please contact support@quantumnous.com
*/

import React, { useEffect, useRef, useState } from 'react';
import { Banner, Button, Card, Col, Form, Row, Spin } from '@douyinfe/semi-ui';
import {
  API,
  removeTrailingSlash,
  showError,
  showSuccess,
} from '../../../helpers';
import { useTranslation } from 'react-i18next';
import { Info } from 'lucide-react';

export default function SettingsPaymentGatewayYima(props) {
  const { t } = useTranslation();
  const sectionTitle = props.hideSectionTitle ? undefined : t('倚码支付设置');
  const [loading, setLoading] = useState(false);
  const [inputs, setInputs] = useState({
    YimaEnabled: false,
    YimaPayAddress: '',
    YimaMerchantId: '',
    YimaMerchantKey: '',
    YimaNotifyUrl: '',
    YimaReturnUrl: '',
    YimaSubscriptionReturnUrl: '',
    YimaAlipayEnabled: true,
    YimaAlipayName: '支付宝 (倚码)',
    YimaWechatEnabled: true,
    YimaWechatName: '微信 (倚码)',
    YimaMinTopUp: 1,
  });
  const formApiRef = useRef(null);

  useEffect(() => {
    if (props.options && formApiRef.current) {
      const currentInputs = {
        YimaEnabled:
          props.options.YimaEnabled === true ||
          props.options.YimaEnabled === 'true',
        YimaPayAddress: props.options.YimaPayAddress || '',
        YimaMerchantId: props.options.YimaMerchantId || '',
        YimaMerchantKey: props.options.YimaMerchantKey || '',
        YimaNotifyUrl: props.options.YimaNotifyUrl || '',
        YimaReturnUrl: props.options.YimaReturnUrl || '',
        YimaSubscriptionReturnUrl:
          props.options.YimaSubscriptionReturnUrl || '',
        YimaAlipayEnabled:
          props.options.YimaAlipayEnabled === undefined
            ? true
            : props.options.YimaAlipayEnabled === true ||
              props.options.YimaAlipayEnabled === 'true',
        YimaAlipayName: props.options.YimaAlipayName || '支付宝 (倚码)',
        YimaWechatEnabled:
          props.options.YimaWechatEnabled === undefined
            ? true
            : props.options.YimaWechatEnabled === true ||
              props.options.YimaWechatEnabled === 'true',
        YimaWechatName: props.options.YimaWechatName || '微信 (倚码)',
        YimaMinTopUp:
          props.options.YimaMinTopUp !== undefined
            ? parseFloat(props.options.YimaMinTopUp)
            : 1,
      };

      setInputs(currentInputs);
      formApiRef.current.setValues(currentInputs);
    }
  }, [props.options]);

  const handleFormChange = (values) => {
    setInputs(values);
  };

  const submitYimaSettings = async () => {
    if (props.options.ServerAddress === '') {
      showError(t('请先填写服务器地址'));
      return;
    }

    setLoading(true);
    try {
      const options = [
        {
          key: 'YimaEnabled',
          value: inputs.YimaEnabled ? 'true' : 'false',
        },
        {
          key: 'YimaPayAddress',
          value: removeTrailingSlash(inputs.YimaPayAddress || ''),
        },
        {
          key: 'YimaMerchantId',
          value: inputs.YimaMerchantId || '',
        },
        {
          key: 'YimaNotifyUrl',
          value: removeTrailingSlash(inputs.YimaNotifyUrl || ''),
        },
        {
          key: 'YimaReturnUrl',
          value: removeTrailingSlash(inputs.YimaReturnUrl || ''),
        },
        {
          key: 'YimaSubscriptionReturnUrl',
          value: removeTrailingSlash(inputs.YimaSubscriptionReturnUrl || ''),
        },
        {
          key: 'YimaAlipayEnabled',
          value: inputs.YimaAlipayEnabled ? 'true' : 'false',
        },
        {
          key: 'YimaAlipayName',
          value: inputs.YimaAlipayName || '',
        },
        {
          key: 'YimaWechatEnabled',
          value: inputs.YimaWechatEnabled ? 'true' : 'false',
        },
        {
          key: 'YimaWechatName',
          value: inputs.YimaWechatName || '',
        },
        {
          key: 'YimaMinTopUp',
          value: inputs.YimaMinTopUp.toString(),
        },
      ];

      if (
        inputs.YimaMerchantKey !== undefined &&
        inputs.YimaMerchantKey !== ''
      ) {
        options.push({ key: 'YimaMerchantKey', value: inputs.YimaMerchantKey });
      }

      const requestQueue = options.map((opt) =>
        API.put('/api/option/', {
          key: opt.key,
          value: opt.value,
        }),
      );

      const results = await Promise.all(requestQueue);

      const errorResults = results.filter((res) => !res.data.success);
      if (errorResults.length > 0) {
        errorResults.forEach((res) => {
          showError(res.data.message);
        });
      } else {
        showSuccess(t('更新成功'));
        props.refresh && props.refresh();
      }
    } catch (error) {
      showError(t('更新失败'));
    }
    setLoading(false);
  };

  return (
    <Spin spinning={loading}>
      <Form
        initValues={inputs}
        onValueChange={handleFormChange}
        getFormApi={(api) => (formApiRef.current = api)}
      >
        <Form.Section text={sectionTitle}>
          <Banner
            type='info'
            icon={<Info size={16} />}
            description={t(
              '当前为倚码支付独立配置页，回调地址留空时将自动使用系统地址。',
            )}
            style={{ marginBottom: 16 }}
          />
          <Row gutter={{ xs: 8, sm: 16, md: 24, lg: 24, xl: 24, xxl: 24 }}>
            <Col xs={24} sm={24} md={6} lg={6} xl={6}>
              <Form.Switch field='YimaEnabled' label={t('启用倚码支付')} />
            </Col>
            <Col xs={24} sm={24} md={6} lg={6} xl={6}>
              <Form.InputNumber
                field='YimaMinTopUp'
                label={t('倚码最低充值数量')}
                placeholder={t('例如：1')}
              />
            </Col>
            <Col xs={24} sm={24} md={12} lg={12} xl={12}>
              <Form.Input
                field='YimaPayAddress'
                label={t('倚码支付地址')}
                placeholder={t('例如：https://zf.rx.sc.cn')}
              />
            </Col>
          </Row>
          <Row
            gutter={{ xs: 8, sm: 16, md: 24, lg: 24, xl: 24, xxl: 24 }}
            style={{ marginTop: 16 }}
          >
            <Col xs={24} sm={24} md={8} lg={8} xl={8}>
              <Form.Input
                field='YimaMerchantId'
                label={t('倚码商户 ID')}
                placeholder={t('例如：10000001')}
              />
            </Col>
            <Col xs={24} sm={24} md={8} lg={8} xl={8}>
              <Form.Input
                field='YimaMerchantKey'
                label={t('倚码商户密钥')}
                placeholder={t('敏感信息不会发送到前端显示')}
                type='password'
              />
            </Col>
            <Col xs={24} sm={24} md={8} lg={8} xl={8}>
              <Form.Input
                field='YimaNotifyUrl'
                label={t('倚码回调地址覆盖')}
                placeholder={t('可选，留空使用默认服务地址')}
              />
            </Col>
          </Row>
          <Row
            gutter={{ xs: 8, sm: 16, md: 24, lg: 24, xl: 24, xxl: 24 }}
            style={{ marginTop: 16 }}
          >
            <Col xs={24} sm={24} md={12} lg={12} xl={12}>
              <Card style={{ height: '100%' }}>
                <Form.Switch
                  field='YimaAlipayEnabled'
                  label={t('启用倚码支付宝')}
                />
                <Form.Input
                  field='YimaAlipayName'
                  label={t('倚码支付宝显示名称')}
                  placeholder={t('例如：支付宝扫码支付')}
                  style={{ marginTop: 16 }}
                />
              </Card>
            </Col>
            <Col xs={24} sm={24} md={12} lg={12} xl={12}>
              <Card style={{ height: '100%' }}>
                <Form.Switch
                  field='YimaWechatEnabled'
                  label={t('启用倚码微信')}
                />
                <Form.Input
                  field='YimaWechatName'
                  label={t('倚码微信显示名称')}
                  placeholder={t('例如：微信扫码支付')}
                  style={{ marginTop: 16 }}
                />
              </Card>
            </Col>
          </Row>
          <Row
            gutter={{ xs: 8, sm: 16, md: 24, lg: 24, xl: 24, xxl: 24 }}
            style={{ marginTop: 16 }}
          >
            <Col xs={24} sm={24} md={12} lg={12} xl={12}>
              <Form.Input
                field='YimaReturnUrl'
                label={t('倚码充值返回地址覆盖')}
                placeholder={t('可选，留空使用默认服务地址')}
              />
            </Col>
            <Col xs={24} sm={24} md={12} lg={12} xl={12}>
              <Form.Input
                field='YimaSubscriptionReturnUrl'
                label={t('倚码订阅返回地址覆盖')}
                placeholder={t('可选，留空使用默认服务地址')}
              />
            </Col>
          </Row>
          <Button onClick={submitYimaSettings} style={{ marginTop: 16 }}>
            {t('更新倚码支付设置')}
          </Button>
        </Form.Section>
      </Form>
    </Spin>
  );
}