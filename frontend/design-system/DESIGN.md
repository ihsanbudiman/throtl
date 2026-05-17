---
version: alpha
name: throtl-xai-dashboard
description: >
  An adaptation of xAI's design language for the Throtl API Gateway dashboard.
  Dark near-black canvas (#0a0a0a), white pill outlines, Geist Mono uppercase labels,
  muted sunset/dusk accent palette reserved for data visualization. Reads as an
  engineering console — precise, unadorned, confident.

colors:
  # ── Brand ──
  primary: "#ffffff"
  on-primary: "#0a0a0a"

  # ── Surfaces ──
  canvas: "#0a0a0a"
  canvas-soft: "#1a1c20"
  canvas-card: "#191919"
  canvas-mid: "#363a3f"
  hairline: "#212327"

  # ── Text ──
  ink: "#ffffff"
  ink-hover: "#fafaf7"
  body: "#dadbdf"
  body-mid: "#7d8187"
  mute: "#7d8187"

  # ── Status (minimal — shape + text first, color second) ──
  success: "#22c55e"
  destructive: "#ef4444"

  # ── Chart Accents (muted palette, used ONLY in data viz) ──
  chart-sunset: "#ff7a17"
  chart-sunset-soft: "#ffc285"
  chart-dusk: "#7c3aed"
  chart-twilight: "#c4b5fd"
  chart-breeze: "#a0c3ec"
  chart-midnight: "#0d1726"

typography:
  display-xl:
    fontFamily: "Geist Variable", Inter, system-ui, sans-serif
    fontSize: 48px
    fontWeight: 500
    lineHeight: 1
    letterSpacing: -1.2px
  display-md:
    fontFamily: "Geist Variable", Inter, system-ui, sans-serif
    fontSize: 32px
    fontWeight: 500
    lineHeight: 1.1
    letterSpacing: -0.8px
  display-sm:
    fontFamily: "Geist Variable", Inter, system-ui, sans-serif
    fontSize: 24px
    fontWeight: 500
    lineHeight: 1.2
    letterSpacing: -0.4px
  display-xs:
    fontFamily: "Geist Variable", Inter, system-ui, sans-serif
    fontSize: 20px
    fontWeight: 500
    lineHeight: 1.3
  body-lg:
    fontFamily: "Geist Variable", Inter, system-ui, sans-serif
    fontSize: 16px
    fontWeight: 400
    lineHeight: 1.5
  body-md:
    fontFamily: "Geist Variable", Inter, system-ui, sans-serif
    fontSize: 14px
    fontWeight: 400
    lineHeight: 1.45
  body-sm:
    fontFamily: "Geist Variable", Inter, system-ui, sans-serif
    fontSize: 13px
    fontWeight: 400
    lineHeight: 1.4
  caption-mono:
    fontFamily: "Geist Mono", ui-monospace, SFMono-Regular, Menlo, Monaco, monospace
    fontSize: 12px
    fontWeight: 400
    lineHeight: 1.3
    letterSpacing: 1.2px
    textTransform: uppercase
  caption-mono-sm:
    fontFamily: "Geist Mono", ui-monospace, SFMono-Regular, Menlo, monospace
    fontSize: 11px
    fontWeight: 400
    lineHeight: 1.2
    letterSpacing: 1px
    textTransform: uppercase
  button-md:
    fontFamily: "Geist Variable", Inter, system-ui, sans-serif
    fontSize: 14px
    fontWeight: 400
    lineHeight: 1

rounded:
  none: 0px
  sm: 8px
  pill: 9999px

spacing:
  xxs: 2px
  xs: 4px
  sm: 8px
  md: 12px
  lg: 16px
  xl: 24px
  2xl: 32px
  3xl: 48px
  4xl: 64px

components:
  button-primary:
    backgroundColor: "{colors.primary}"
    textColor: "{colors.on-primary}"
    borderColor: "{colors.primary}"
    typography: "{typography.button-md}"
    rounded: "{rounded.pill}"
    padding: "{spacing.xs} {spacing.md}"
  button-outline:
    backgroundColor: transparent
    textColor: "{colors.ink}"
    borderColor: "{colors.hairline}"
    typography: "{typography.button-md}"
    rounded: "{rounded.pill}"
    padding: "{spacing.xs} {spacing.md}"
  button-ghost:
    backgroundColor: "{colors.canvas}"
    textColor: "{colors.body}"
    borderColor: transparent
    typography: "{typography.button-md}"
    rounded: "{rounded.pill}"
    padding: "{spacing.xs} {spacing.md}"
  text-input:
    backgroundColor: "{colors.canvas-soft}"
    textColor: "{colors.ink}"
    borderColor: "{colors.hairline}"
    typography: "{typography.body-md}"
    rounded: "{rounded.sm}"
    padding: "{spacing.md} {spacing.lg}"
  card:
    backgroundColor: "{colors.canvas-card}"
    textColor: "{colors.ink}"
    borderColor: "{colors.hairline}"
    typography: "{typography.body-md}"
    rounded: "{rounded.sm}"
    padding: "{spacing.xl}"
  dialog:
    backgroundColor: "{colors.canvas}"
    textColor: "{colors.ink}"
    borderColor: "{colors.hairline}"
    rounded: "{rounded.sm}"
    padding: "{spacing.xl}"
  sidebar:
    backgroundColor: "{colors.canvas}"
    textColor: "{colors.body}"
    width: 240px
    navItemPadding: "{spacing.sm} {spacing.lg}"
    activeIndicator: "{colors.primary}"
  table:
    headerBackground: "{colors.canvas}"
    headerTypography: "{typography.caption-mono}"
    bodyTypography: "{typography.body-sm}"
    cellPadding: "{spacing.md} {spacing.lg}"
    rowBorder: "{colors.hairline}"
  badge:
    backgroundColor: "{colors.canvas-soft}"
    textColor: "{colors.body}"
    rounded: "{rounded.pill}"
    padding: "{spacing.xxs} {spacing.sm}"
  badge-success:
    backgroundColor: "{colors.success}20"
    textColor: "{colors.success}"
    rounded: "{rounded.pill}"
    padding: "{spacing.xxs} {spacing.sm}"
  badge-destructive:
    backgroundColor: "{colors.destructive}20"
    textColor: "{colors.destructive}"
    rounded: "{rounded.pill}"
    padding: "{spacing.xxs} {spacing.sm}"
  toast:
    backgroundColor: "{colors.canvas-card}"
    textColor: "{colors.ink}"
    borderColor: "{colors.hairline}"
    rounded: "{rounded.sm}"
    padding: "{spacing.md} {spacing.lg}"
    typography: "{typography.body-sm}"
  switch:
    backgroundColor: "{colors.canvas-mid}"
    checkedBackgroundColor: "{colors.primary}"
    thumbColor: "{colors.canvas}"
    rounded: "{rounded.pill}"
  divider:
    borderColor: "{colors.hairline}"
  skeleton:
    backgroundColor: "{colors.canvas-soft}"
    rounded: "{rounded.sm}"
  empty-state:
    backgroundColor: "{colors.canvas-soft}"
    rounded: "{rounded.sm}"
    padding: "{spacing.3xl}"
---
