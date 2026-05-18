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

import React from 'react';
import { Button } from '@douyinfe/semi-ui';
import PricingGroups from '../filter/PricingGroups';
import PricingQuotaTypes from '../filter/PricingQuotaTypes';
import PricingEndpointTypes from '../filter/PricingEndpointTypes';
import PricingVendors from '../filter/PricingVendors';
import PricingTags from '../filter/PricingTags';

import { getLobeHubIcon } from '../../../../helpers';
import { resetPricingFilters } from '../../../../helpers/utils';
import { usePricingFilterCounts } from '../../../../hooks/model-pricing/usePricingFilterCounts';

const ModelsVendorList = ({ filterVendor, setFilterVendor, models = [], t }) => {
  const vendors = React.useMemo(() => {
    const vendorMap = new Map();
    let hasUnknownVendor = false;

    models.forEach((model) => {
      if (model.vendor_name) {
        if (!vendorMap.has(model.vendor_name)) {
          vendorMap.set(model.vendor_name, {
            name: model.vendor_name,
            icon: model.vendor_icon,
          });
        }
      } else {
        hasUnknownVendor = true;
      }
    });

    const vendorList = Array.from(vendorMap.values()).sort((a, b) =>
      a.name.localeCompare(b.name),
    );

    if (hasUnknownVendor) {
      vendorList.push({ name: 'unknown', icon: null, label: t('未知供应商') });
    }

    return vendorList;
  }, [models, t]);

  const handleSelectVendor = (vendorName) => {
    setFilterVendor(filterVendor === vendorName ? 'all' : vendorName);
  };

  return (
    <div className='models-vendor-filter'>
      <div className='models-filter-title'>{t('筛选')}</div>
      <div className='models-filter-subtitle'>{t('系列 / 厂商')}</div>
      <div className='models-vendor-list'>
        {vendors.map((vendor) => {
          const isActive = filterVendor === vendor.name;
          const label = vendor.label || vendor.name;

          return (
            <button
              type='button'
              key={vendor.name}
              className={`models-vendor-item ${isActive ? 'active' : ''}`}
              onClick={() => handleSelectVendor(vendor.name)}
              title={label}
            >
              <span className='models-vendor-icon'>
                {vendor.icon ? getLobeHubIcon(vendor.icon, 22) : label.charAt(0)}
              </span>
              <span className='models-vendor-name'>{label}</span>
            </button>
          );
        })}
      </div>
    </div>
  );
};

const PricingSidebar = ({
  showWithRecharge,
  setShowWithRecharge,
  currency,
  setCurrency,
  handleChange,
  setActiveKey,
  showRatio,
  setShowRatio,
  viewMode,
  setViewMode,
  filterGroup,
  setFilterGroup,
  handleGroupClick,
  filterQuotaType,
  setFilterQuotaType,
  filterEndpointType,
  setFilterEndpointType,
  filterVendor,
  setFilterVendor,
  filterTag,
  setFilterTag,
  currentPage,
  setCurrentPage,
  tokenUnit,
  setTokenUnit,
  loading,
  modelListMode = false,
  t,
  ...categoryProps
}) => {
  const {
    quotaTypeModels,
    endpointTypeModels,
    vendorModels,
    tagModels,
    groupCountModels,
  } = usePricingFilterCounts({
    models: categoryProps.models,
    filterGroup,
    filterQuotaType,
    filterEndpointType,
    filterVendor,
    filterTag,
    searchValue: categoryProps.searchValue,
  });

  const handleResetFilters = () => {
    if (modelListMode) {
      setFilterVendor?.('all');
      setCurrentPage?.(1);
      return;
    }

    resetPricingFilters({
      handleChange,
      setShowWithRecharge,
      setCurrency,
      setShowRatio,
      setViewMode,
      setFilterGroup,
      setFilterQuotaType,
      setFilterEndpointType,
      setFilterVendor,
      setFilterTag,
      setCurrentPage,
      setTokenUnit,
    });
  };

  return (
    <div className='p-2'>
      {modelListMode ? (
        <ModelsVendorList
          filterVendor={filterVendor}
          setFilterVendor={setFilterVendor}
          models={categoryProps.models}
          t={t}
        />
      ) : (
        <>
      <div className='flex items-center justify-between mb-6'>
        <div className='text-lg font-semibold text-gray-800'>{t('筛选')}</div>
        <Button
          theme='outline'
          type='tertiary'
          onClick={handleResetFilters}
          className='text-gray-500 hover:text-gray-700'
        >
          {t('重置')}
        </Button>
      </div>

      <PricingVendors
        filterVendor={filterVendor}
        setFilterVendor={setFilterVendor}
        models={vendorModels}
        allModels={categoryProps.models}
        loading={loading}
        t={t}
      />

          <PricingGroups
            filterGroup={filterGroup}
            setFilterGroup={handleGroupClick}
            usableGroup={categoryProps.usableGroup}
            groupRatio={categoryProps.groupRatio}
            models={groupCountModels}
            loading={loading}
            t={t}
          />

          <PricingQuotaTypes
            filterQuotaType={filterQuotaType}
            setFilterQuotaType={setFilterQuotaType}
            models={quotaTypeModels}
            loading={loading}
            t={t}
          />

          <PricingTags
            filterTag={filterTag}
            setFilterTag={setFilterTag}
            models={tagModels}
            allModels={categoryProps.models}
            loading={loading}
            t={t}
          />

          <PricingEndpointTypes
            filterEndpointType={filterEndpointType}
            setFilterEndpointType={setFilterEndpointType}
            models={endpointTypeModels}
            allModels={categoryProps.models}
            loading={loading}
            t={t}
          />
        </>
      )}
    </div>
  );
};

export default PricingSidebar;
