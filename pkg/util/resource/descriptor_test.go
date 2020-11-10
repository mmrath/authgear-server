package resource_test

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/spf13/afero"

	"github.com/authgear/authgear-server/pkg/util/resource"
)

func TestSimpleDescriptor(t *testing.T) {
	Convey("SimpleDescriptor", t, func() {
		fsA := afero.NewMemMapFs()
		fsB := afero.NewMemMapFs()
		r := &resource.Registry{}
		manager := resource.NewManager(r, []resource.Fs{
			resource.AferoFs{Fs: fsA},
			resource.AferoFs{Fs: fsB, IsAppFs: true},
		})

		simple := resource.SimpleDescriptor{
			Path: "static/somefile.txt",
		}
		r.Register(simple)

		writeFile := func(fs afero.Fs, data string) {
			_ = fs.MkdirAll("static/", 0777)
			_ = afero.WriteFile(fs, "static/somefile.txt", []byte(data), 0666)
		}

		read := func(view resource.View) (str string, err error) {
			result, err := manager.Read(simple, view)
			if err != nil {
				return
			}
			bytes := result.([]byte)
			str = string(bytes)
			return
		}

		Convey("AppFileView: not found", func() {
			writeFile(fsA, "file in non-app FS")

			_, err := read(resource.AppFile{})
			So(err, ShouldBeError, "specified resource is not configured")
		})

		Convey("AppFileView: found", func() {
			writeFile(fsB, "file in app FS")

			data, err := read(resource.AppFile{})
			So(err, ShouldBeNil)
			So(data, ShouldEqual, "file in app FS")
		})

		Convey("EffectiveFileView", func() {
			writeFile(fsA, "file in non-app FS")
			writeFile(fsB, "file in app FS")

			data, err := read(resource.EffectiveFile{})
			So(err, ShouldBeNil)
			So(data, ShouldEqual, "file in app FS")
		})
	})
}
