import React from 'react';

import { NavModelItem, PageLayoutType } from '@grafana/data';

export interface PluginPageProps {
  /** Shown under main heading */
  subTitle?: React.ReactNode;
  pageNav?: NavModelItem;
  children: React.ReactNode;
  layout?: PageLayoutType;
}

export type PluginPageType = React.ComponentType<PluginPageProps>;

export let PluginPage: PluginPageType = ({ children }) => {
  return <div>{children}</div>;
};

/**
 * Used to bootstrap the PluginPage during application start
 * is exposed via runtime.
 *
 * @internal
 */
export function setPluginPage(component: PluginPageType) {
  PluginPage = component;
}
