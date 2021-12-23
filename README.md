Deskpot - React + Go
--------------------

Create desktop application UI using React and native webview binding for Go. 
This is a tool to help you develop desktop apps faster by scaffolding
and providing commands that can run, package your application.

> See also

[Native Webview Wrapper](https://github.com/webview/webview)

### Development Status

| Feature | Mac | Windows | Linux |
|---------|-----|---------|-------|
| Webview binding | ✅ | ✅ | ✅ |
| Hot reloading: `dpot run` | ✅ | ❌ | ✅ |
| Packaging: `dpot pack` | ✅ | ❌ | ❌ |
| Menu binding | ❌ | ❌ | ❌ |
| Notifications | ❌ | ❌ | ❌ |
| Tray support | ❌ | ❌ | ❌ |


## Installing

```bash
go get -u github.com/codekidX/deskpot/cmd/dpot
```

## Getting Started

Make sure you have NPM installed.

#### Create

```
dpot new myapp
```

Once the scaffolding is done, run this command to start with your development

#### Develop

```
dpot run
```

This will start your `webpack-dev-server` with the webview application. This
will hot-reload your changes.

#### Package

Supported currently only for **mac**.

```
dpot pack mac
```

This will create your Deskpot developed app inside `out/1.0/myapp`. You can
drag and drop this to your `Applications` folder.

## Configuration

The configuration file for common Deskpot variables are stored inside
`deskpot.json`. You can change this as per your requirements. 

> See also

[Mac App Categories](https://developer.apple.com/documentation/bundleresources/information_property_list/lsapplicationcategorytype)

The `icon` will use default deskpot icon or you can specify your own `Icon.icns`
path.

```json
{
    "identifier": "com.deskpot.myapp",
    "name": "Myapp",
    "description": "This application is created with Deskpot",
    "version": 1.0,
    "run_id": "123781212",
    "osx_category": "public.app-category.developer-tools",
    "publish": {
        "icon": "DEFAULT",
        "copyright": {
            "year": "2021",
            "name": "Deskpot Owner"
        }
    }
}
```

