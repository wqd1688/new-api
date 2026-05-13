/*
Copyright (C) 2023-2026 QuantumNous

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
import { useEffect, useState } from 'react'
import { useForm } from 'react-hook-form'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Switch } from '@/components/ui/switch'
import { SettingsSection } from '../components/settings-section'
import { useUpdateOption } from '../hooks/use-update-option'
import { removeTrailingSlash } from './utils'

export interface YimaSettingsValues {
  YimaEnabled: boolean
  YimaPayAddress: string
  YimaMerchantId: string
  YimaMerchantKey: string
  YimaNotifyUrl: string
  YimaReturnUrl: string
  YimaSubscriptionReturnUrl: string
  YimaAlipayEnabled: boolean
  YimaAlipayName: string
  YimaWechatEnabled: boolean
  YimaWechatName: string
  YimaMinTopUp: number
}

interface Props {
  defaultValues: YimaSettingsValues
}

export function YimaSettingsSection(props: Props) {
  const { t } = useTranslation()
  const updateOption = useUpdateOption()
  const [loading, setLoading] = useState(false)

  const form = useForm<YimaSettingsValues>({
    defaultValues: props.defaultValues,
  })

  useEffect(() => {
    form.reset(props.defaultValues)
  }, [props.defaultValues, form])

  const handleSave = async () => {
    const values = form.getValues()

    setLoading(true)
    try {
      const options: Array<{ key: string; value: string }> = [
        { key: 'YimaEnabled', value: values.YimaEnabled ? 'true' : 'false' },
        {
          key: 'YimaPayAddress',
          value: removeTrailingSlash(values.YimaPayAddress || ''),
        },
        { key: 'YimaMerchantId', value: values.YimaMerchantId || '' },
        {
          key: 'YimaNotifyUrl',
          value: removeTrailingSlash(values.YimaNotifyUrl || ''),
        },
        {
          key: 'YimaReturnUrl',
          value: removeTrailingSlash(values.YimaReturnUrl || ''),
        },
        {
          key: 'YimaSubscriptionReturnUrl',
          value: removeTrailingSlash(values.YimaSubscriptionReturnUrl || ''),
        },
        {
          key: 'YimaAlipayEnabled',
          value: values.YimaAlipayEnabled ? 'true' : 'false',
        },
        {
          key: 'YimaAlipayName',
          value: values.YimaAlipayName || '',
        },
        {
          key: 'YimaWechatEnabled',
          value: values.YimaWechatEnabled ? 'true' : 'false',
        },
        {
          key: 'YimaWechatName',
          value: values.YimaWechatName || '',
        },
        {
          key: 'YimaMinTopUp',
          value: String(values.YimaMinTopUp ?? 1),
        },
      ]

      if ((values.YimaMerchantKey || '').trim()) {
        options.push({
          key: 'YimaMerchantKey',
          value: values.YimaMerchantKey,
        })
      }

      for (const option of options) {
        await updateOption.mutateAsync(option)
      }

      toast.success(t('Updated successfully'))
    } catch {
      toast.error(t('Update failed'))
    } finally {
      setLoading(false)
    }
  }

  return (
    <SettingsSection
      title={t('Yima Gateway')}
      description={t('Configuration for Yima merchant OpenAPI integration')}
    >
      <div className='grid grid-cols-1 gap-4 md:grid-cols-2'>
        <div className='flex items-center gap-2 rounded-lg border p-4 md:col-span-2'>
          <Switch
            checked={!!form.watch('YimaEnabled')}
            onCheckedChange={(value) => form.setValue('YimaEnabled', value)}
          />
          <div>
            <Label>{t('Enable Yima')}</Label>
            <p className='text-muted-foreground mt-1 text-sm'>
              {t('Expose WeChat and Alipay payment buttons through Yima')}
            </p>
          </div>
        </div>

        <div className='grid gap-1.5'>
          <Label>{t('Yima endpoint')}</Label>
          <Input
            placeholder={t('https://zf.rx.sc.cn')}
            {...form.register('YimaPayAddress')}
          />
          <p className='text-muted-foreground text-xs'>
            {t('Base domain of your Yima payment service')}
          </p>
        </div>

        <div className='grid gap-1.5'>
          <Label>{t('Yima merchant ID')}</Label>
          <Input placeholder='10000001' {...form.register('YimaMerchantId')} />
        </div>

        <div className='grid gap-1.5'>
          <Label>{t('Yima merchant key')}</Label>
          <Input
            type='password'
            placeholder={t('Enter new key to update')}
            autoComplete='new-password'
            {...form.register('YimaMerchantKey')}
          />
          <p className='text-muted-foreground text-xs'>
            {t('Leave blank unless rotating the secret')}
          </p>
        </div>

        <div className='grid gap-1.5'>
          <Label>{t('Yima minimum top-up')}</Label>
          <Input
            type='number'
            min={0}
            step='1'
            {...form.register('YimaMinTopUp', { valueAsNumber: true })}
          />
          <p className='text-muted-foreground text-xs'>
            {t('Minimum amount displayed for Yima payment buttons')}
          </p>
        </div>

        <div className='grid gap-4 rounded-lg border p-4 md:col-span-2'>
          <div>
            <Label>{t('Yima payment methods')}</Label>
            <p className='text-muted-foreground mt-1 text-sm'>
              {t('Configure Alipay and WeChat independently')}
            </p>
          </div>

          <div className='grid grid-cols-1 gap-4 md:grid-cols-2'>
            <div className='space-y-3 rounded-lg border p-4'>
              <div className='flex items-center gap-2'>
                <Switch
                  checked={!!form.watch('YimaAlipayEnabled')}
                  onCheckedChange={(value) =>
                    form.setValue('YimaAlipayEnabled', value)
                  }
                />
                <Label>{t('Enable Yima Alipay')}</Label>
              </div>
              <div className='grid gap-1.5'>
                <Label>{t('Yima Alipay display name')}</Label>
                <Input {...form.register('YimaAlipayName')} />
                <p className='text-muted-foreground text-xs'>
                  {t('Shown on the recharge page')}
                </p>
              </div>
            </div>

            <div className='space-y-3 rounded-lg border p-4'>
              <div className='flex items-center gap-2'>
                <Switch
                  checked={!!form.watch('YimaWechatEnabled')}
                  onCheckedChange={(value) =>
                    form.setValue('YimaWechatEnabled', value)
                  }
                />
                <Label>{t('Enable Yima WeChat')}</Label>
              </div>
              <div className='grid gap-1.5'>
                <Label>{t('Yima WeChat display name')}</Label>
                <Input {...form.register('YimaWechatName')} />
                <p className='text-muted-foreground text-xs'>
                  {t('Shown on the recharge page')}
                </p>
              </div>
            </div>
          </div>
        </div>

        <div className='grid gap-1.5'>
          <Label>{t('Yima notify URL')}</Label>
          <Input
            placeholder={t('Optional override')}
            {...form.register('YimaNotifyUrl')}
          />
        </div>

        <div className='grid gap-1.5'>
          <Label>{t('Yima top-up return URL')}</Label>
          <Input
            placeholder={t('Optional override')}
            {...form.register('YimaReturnUrl')}
          />
        </div>

        <div className='grid gap-1.5 md:col-span-2'>
          <Label>{t('Yima subscription return URL')}</Label>
          <Input
            placeholder={t('Optional override')}
            {...form.register('YimaSubscriptionReturnUrl')}
          />
        </div>
      </div>

      <Button type='button' onClick={handleSave} disabled={loading}>
        {loading ? t('Saving...') : t('Save Yima settings')}
      </Button>
    </SettingsSection>
  )
}