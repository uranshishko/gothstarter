# GOTH starter template

Get started by cloning the project from [github.com/uranshishko/gothstarter]

Make sure to:

- Install [Air](https://github.com/air-verse/air) for live reloading.
- Install [TEMPL](https://github.com/a-h/templ)
- Run `npm install` to install dev dependencies (daisyUI and tailwindcss)

## DEV

Run folloring commands in separate terminals:

```bash
air
```

```bash
npx tailwindcss -i ./views/css/app.css -o ./public/styles.css --watch
```

```bash
templ generate --watch
```

## BUILD

Run `make build` to build the project, or `make`to build and run it.
