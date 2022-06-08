---
title: Getting Started
aside: Docs
---

# Getting Started
As of now, the easiest way to install Forge is with Go's `install` command. [Install Go](https://go.dev/doc/install) if you haven't already.

## Install Forge
```bash
go install github.com/twharmon/forge@latest
```

## Create a New Site
```bash
forge new my-site
```

## Add a Theme
```bash
cd my-site
forge add-theme https://github.com/twharmon/forge-dox
```
Reference the theme's readme for information on how to use it.

## Start the Development Server
```bash
forge serve
```

## Build for Production
```bash
forge build
```
