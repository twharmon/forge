# Forge Dox
Simple theme for Forge.

## Example Config
```yml
DevServer:
  Port: 8000

Markdown:
  Extensions:
    - footnote
    - linkify

Theme:
  Name: forge-dox
  Params:
    SiteName: My Site
    SiteLogo: /logo.png
    CodeSyntax: true
    Menu:
      Main:
        - Name: Introduction
          Url: /
        - Name: Docs
          Url: /docs/getting-started
        - Name: Github
          Url: https://github.com/example/example
      Docs:
        - Name: Getting Started
          Url: /docs/getting-started
        - Name: Components
          Url: /docs/components

```
