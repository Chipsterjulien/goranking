# Goranking
It move, search duplicate, sort pictures/movies into a structure year/month directories

## Dependencies
* github.com/spf13/cobra
* github.com/rwcarlsen/goexif/exif
* github.com/mitchellh/go-homedir
* github.com/spf13/viper

## Install dependencies

```
for dep in "github.com/spf13/cobra" "github.com/rwcarlsen/goexif/exif" "github.com/mitchellh/go-homedir" "github.com/spf13/viper"
do
  go get -u -v "$dep"
done
```

## Compile

```
go build
```
