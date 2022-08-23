# goconvert
A simple, web-based file conversion tool written in Go!

### How it works
goconvert uses the built-in Go `net/http` package to spin up a simple but responsive http server. From here the user can navigate to `localhost:4433` - here they are greeted with a simple, easy-to-use interface.

![](docs/images/homescreen.png)

From here the user can choose to upload a document (only PNG/JPG/PDF files are currently supported), the user then selects what format they want to convert to (again, only PNG/JPG/PDF at the moment) and the click `Upload`.

![](docs/images/interface.png)

After a very small wait, the resulting image is automatically downloaded (or the user is prompted to select somewhere to save it, depending on the browser).

![](docs/images/autodownload.png)

Once the user has downloaded the image (or are using a browser that doesn't support the auto-download feature), they are presented with a visual representation of the image. This is so that, should the user need/want to, they can download the image using the traditional `Right-click > Save as` method.

![](docs/images/results.png)