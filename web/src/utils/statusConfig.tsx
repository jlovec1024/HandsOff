import {
  CheckCircleOutlined,
  CloseCircleOutlined,
  ClockCircleOutlined,
} from "@ant-design/icons";
import { Tag } from "antd";
import React from "react";

/**
 * Centralized status configuration for review status display
 * Eliminates code duplication across List, Detail, and Dashboard components
 */

export type ReviewStatus = "completed" | "failed" | "processing" | "pending";

export interface StatusConfig {
  color: string;
  icon: React.ReactNode;
  text: string;
  tagColor: string;
}

export const STATUS_CONFIG: Record<ReviewStatus, StatusConfig> = {
  completed: {
    color: "#52c41a",
    icon: <CheckCircleOutlined style={{ color: "#52c41a" }} />,
    text: "Completed",
    tagColor: "success",
  },
  failed: {
    color: "#ff4d4f",
    icon: <CloseCircleOutlined style={{ color: "#ff4d4f" }} />,
    text: "Failed",
    tagColor: "error",
  },
  processing: {
    color: "#1890ff",
    icon: <ClockCircleOutlined style={{ color: "#1890ff" }} spin />,
    text: "Processing",
    tagColor: "processing",
  },
  pending: {
    color: "#d9d9d9",
    icon: <ClockCircleOutlined style={{ color: "#d9d9d9" }} />,
    text: "Pending",
    tagColor: "default",
  },
};

/**
 * Get status configuration with fallback to pending
 */
export const getStatusConfig = (status: string): StatusConfig => {
  return STATUS_CONFIG[status as ReviewStatus] || STATUS_CONFIG.pending;
};

/**
 * Render a status tag (for tables and detail views)
 */
export const renderStatusTag = (status: string): React.ReactNode => {
  const config = getStatusConfig(status);
  return (
    <Tag color={config.tagColor} icon={config.icon}>
      {config.text}
    </Tag>
  );
};

/**
 * Render status with icon and text (for inline display)
 */
export const renderStatusInline = (status: string): React.ReactNode => {
  const config = getStatusConfig(status);
  return (
    <span style={{ display: "inline-flex", alignItems: "center", gap: 4 }}>
      {config.icon}
      <span style={{ textTransform: "capitalize" }}>{status}</span>
    </span>
  );
};

/**
 * Score color helper - centralized score color logic
 */
export const getScoreColor = (score: number): string => {
  if (score >= 80) return "#52c41a";
  if (score >= 60) return "#faad14";
  return "#ff4d4f";
};

/**
 * Format token count for display (e.g., 1234 -> "1.2k", 1234567 -> "1.2M")
 */
export const formatTokens = (tokens: number): string => {
  if (tokens >= 1000000) return `${(tokens / 1000000).toFixed(1)}M`;
  if (tokens >= 1000) return `${(tokens / 1000).toFixed(1)}k`;
  return tokens.toString();
};

/**
 * Format duration in ms to human readable (e.g., 1500 -> "1.50s", 500 -> "500ms")
 */
export const formatDuration = (ms: number): string => {
  if (ms >= 1000) return `${(ms / 1000).toFixed(2)}s`;
  return `${ms}ms`;
};
