import { size } from 'lodash';
import React, { useCallback, useState } from 'react';

import { SelectableValue, toOption } from '@grafana/data';
import { InlineLabel, Select, InlineFormLabel, InlineSwitch, Icon } from '@grafana/ui';

import { OpenTsdbFilter, OpenTsdbQuery } from '../types';

export interface FilterSectionProps {
  query: OpenTsdbQuery;
  onChange: (query: OpenTsdbQuery) => void;
  onRunQuery: () => void;
  suggestTagKeys: (query: OpenTsdbQuery) => Promise<string[]>;
  filterTypes: string[];
  suggestTagValues: () => Promise<SelectableValue[]>;
}

export function FilterSection({
  query,
  onChange,
  onRunQuery,
  suggestTagKeys,
  filterTypes,
  suggestTagValues,
}: FilterSectionProps) {
  const [tagKeys, updTagKeys] = useState<Array<SelectableValue<string>>>();
  const [keyIsLoading, updKeyIsLoading] = useState<boolean>();

  const [tagValues, updTagValues] = useState<Array<SelectableValue<string>>>();
  const [valueIsLoading, updValueIsLoading] = useState<boolean>();

  const [addFilterMode, updAddFilterMode] = useState<boolean>(false);

  const [curFilterType, updCurFilterType] = useState<string>('iliteral_or');
  const [curFilterKey, updCurFilterKey] = useState<string>('');
  const [curFilterValue, updCurFilterValue] = useState<string>('');
  const [curFilterGroupBy, updCurFilterGroupBy] = useState<boolean>(false);

  const [errors, setErrors] = useState<string>('');

  const filterTypesOptions = filterTypes.map((value: string) => toOption(value));

  function changeAddFilterMode() {
    updAddFilterMode(!addFilterMode);
  }

  function addFilter() {
    if (query.tags && size(query.tags) > 0) {
      const err = 'Please remove tags to use filters, tags and filters are mutually exclusive.';
      setErrors(err);
      return;
    }

    if (!addFilterMode) {
      updAddFilterMode(true);
      return;
    }

    // Add the filter to the query
    const currentFilter = {
      type: curFilterType,
      tagk: curFilterKey,
      filter: curFilterValue,
      groupBy: curFilterGroupBy,
    };

    // filters may be undefined
    query.filters = query.filters ? query.filters.concat([currentFilter]) : [currentFilter];

    // reset the inputs
    updCurFilterType('literal_or');
    updCurFilterKey('');
    updCurFilterValue('');
    updCurFilterGroupBy(false);

    // fire the query
    onChange(query);
    onRunQuery();

    // close the filter ditor
    changeAddFilterMode();
  }

  function removeFilter(index: number) {
    query.filters?.splice(index, 1);
    // fire the query
    onChange(query);
    onRunQuery();
  }

  function editFilter(fil: OpenTsdbFilter, idx: number) {
    removeFilter(idx);
    updCurFilterKey(fil.tagk);
    updCurFilterValue(fil.filter);
    updCurFilterType(fil.type);
    updCurFilterGroupBy(fil.groupBy);
    addFilter();
  }

  // We are matching words split with space
  const splitSeparator = ' ';
  const customFilterOption = useCallback((option: SelectableValue<string>, searchQuery: string) => {
    const label = option.value ?? '';

    const searchWords = searchQuery.split(splitSeparator);
    return searchWords.reduce((acc, cur) => acc && label.toLowerCase().includes(cur.toLowerCase()), true);
  }, []);

  return (
    <div className="gf-form-inline" data-testid={testIds.section}>
      <div className="gf-form">
        <InlineFormLabel
          className="query-keyword"
          width={8}
          tooltip={<div>Filters does not work with tags, either of the two will work but not both.</div>}
        >
          Filters
        </InlineFormLabel>
        {query.filters &&
          query.filters.map((fil: OpenTsdbFilter, idx: number) => {
            return (
              <InlineFormLabel key={idx} width="auto" data-testid={testIds.list + idx}>
                {fil.tagk} = {fil.type}({fil.filter}), groupBy = {'' + fil.groupBy}
                <a onClick={() => editFilter(fil, idx)}>
                  <Icon name={'pen'} />
                </a>
                <a onClick={() => removeFilter(idx)} data-testid={testIds.remove}>
                  <Icon name={'times'} />
                </a>
              </InlineFormLabel>
            );
          })}
        {!addFilterMode && (
          <label className="gf-form-label query-keyword">
            <a onClick={changeAddFilterMode} data-testid={testIds.open}>
              <Icon name={'plus'} />
            </a>
          </label>
        )}
      </div>
      {addFilterMode && (
        <div className="gf-form-inline">
          <div className="gf-form">
            <Select
              inputId="opentsdb-suggested-tagk-select"
              className="gf-form-input"
              value={curFilterKey ? toOption(curFilterKey) : undefined}
              placeholder="key"
              onOpenMenu={async () => {
                updKeyIsLoading(true);
                const tKs = await suggestTagKeys(query);
                const tKsOptions = tKs.map((value: string) => toOption(value));
                updTagKeys(tKsOptions);
                updKeyIsLoading(false);
              }}
              isLoading={keyIsLoading}
              options={tagKeys}
              onChange={({ value }) => {
                if (value) {
                  updCurFilterKey(value);
                }
              }}
            />
          </div>

          <div className="gf-form">
            <InlineLabel className="width-4 query-keyword">Type</InlineLabel>
            <Select
              inputId="opentsdb-aggregator-select"
              value={curFilterType ? toOption(curFilterType) : undefined}
              options={filterTypesOptions}
              onChange={({ value }) => {
                if (value) {
                  updCurFilterType(value);
                }
              }}
            />
          </div>

          <div className="gf-form">
            <Select
              inputId="opentsdb-suggested-tagv-select"
              className="gf-form-input"
              value={curFilterValue ? toOption(curFilterValue) : undefined}
              placeholder="filter"
              allowCustomValue
              filterOption={customFilterOption}
              onOpenMenu={async () => {
                if (!tagValues) {
                  updValueIsLoading(true);
                  const tVs = await suggestTagValues();
                  updTagValues(tVs);
                  updValueIsLoading(false);
                }
              }}
              isLoading={valueIsLoading}
              options={tagValues}
              onChange={({ value }) => {
                if (value) {
                  updCurFilterValue(value);
                }
              }}
            />
          </div>

          <InlineFormLabel width={5} className="query-keyword">
            Group by
          </InlineFormLabel>
          <InlineSwitch
            value={curFilterGroupBy}
            onChange={() => {
              // DO NOT RUN THE QUERY HERE
              // OLD FUNCTIONALITY RAN THE QUERY
              updCurFilterGroupBy(!curFilterGroupBy);
            }}
          />
          <div className="gf-form">
            {errors && (
              <label className="gf-form-label" title={errors} data-testid={testIds.error}>
                <Icon name={'exclamation-triangle'} color={'rgb(229, 189, 28)'} />
              </label>
            )}

            <label className="gf-form-label">
              <a onClick={addFilter}>add filter</a>
              <a onClick={changeAddFilterMode}>
                <Icon name={'times'} />
              </a>
            </label>
          </div>
        </div>
      )}
      <div className="gf-form gf-form--grow">
        <div className="gf-form-label gf-form-label--grow"></div>
      </div>
    </div>
  );
}

export const testIds = {
  section: 'opentsdb-filter',
  open: 'opentsdb-filter-editor',
  list: 'opentsdb-filter-list',
  error: 'opentsdb-filter-error',
  remove: 'opentsdb-filter-remove',
};
