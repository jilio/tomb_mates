package tinyrpg

import (
	"bytes"
	"image"
	"image/png"

	"github.com/gobuffalo/packr/v2"
)

func LoadResources() (map[string][]image.Image, error) {
	images := map[string]image.Image{}
	sprites := map[string][]image.Image{}
	imagesBox := packr.New("images", "./resources/sprites")

	list := imagesBox.List()
	for _, filename := range list {
		data, err := imagesBox.Find(filename)
		if err != nil {
			return sprites, err
		}
		img, err := png.Decode(bytes.NewReader(data))
		if err != nil {
			return sprites, err
		}

		images[filename] = img
	}

	sprites["big_demon_idle"] = []image.Image{
		images["big_demon_idle_anim_f0.png"],
		images["big_demon_idle_anim_f1.png"],
		images["big_demon_idle_anim_f2.png"],
		images["big_demon_idle_anim_f3.png"],
	}
	sprites["big_demon_run"] = []image.Image{
		images["big_demon_run_anim_f0.png"],
		images["big_demon_run_anim_f1.png"],
		images["big_demon_run_anim_f2.png"],
		images["big_demon_run_anim_f3.png"],
	}
	sprites["big_zombie_idle"] = []image.Image{
		images["big_zombie_idle_anim_f0.png"],
		images["big_zombie_idle_anim_f1.png"],
		images["big_zombie_idle_anim_f2.png"],
		images["big_zombie_idle_anim_f3.png"],
	}
	sprites["big_zombie_run"] = []image.Image{
		images["big_zombie_run_anim_f0.png"],
		images["big_zombie_run_anim_f1.png"],
		images["big_zombie_run_anim_f2.png"],
		images["big_zombie_run_anim_f3.png"],
	}

	return sprites, nil
}
