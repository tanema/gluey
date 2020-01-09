package main

import (
	"errors"
	"time"

	"github.com/tanema/gluey"
)

func main() {
	ctx := gluey.New()
	ctx.AskDefault("Username", "foo")
	ctx.Password("Password")
	ctx.Confirm("Skip Run?", true)
	ctx.SelectMultiple("What's your text editor", []string{"Vim", "Emacs", "Sublime", "VSCode", "Atom", "other", "Vim", "Emacs", "Sublime", "VSCode", "Atom", "other"})
	ctx.InMeasuredFrame("Build", func(c *gluey.Ctx, f *gluey.Frame) error {
		c.InFrame("Cloning", func(c *gluey.Ctx, f *gluey.Frame) error {
			return c.Progress(100, func(c *gluey.Ctx, bar *gluey.Bar) error {
				for i := 1; i <= 100; i++ {
					bar.Tick(1)
					time.Sleep(1 * time.Millisecond)
				}
				return nil
			})
		})

		c.Println("Requesting something")
		f.Divider("Request Failed", "yellow")
		c.Println("https failed or something")

		pgroup := c.NewProgressGroup()
		pgroup.Go("Git Clone", 100, func(c *gluey.Ctx, bar *gluey.Bar) error {
			for i := 1; i <= 100; i++ {
				bar.Tick(1)
				time.Sleep(1 * time.Millisecond)
			}
			return nil
		})
		pgroup.Go("Docker Image", 200, func(c *gluey.Ctx, bar *gluey.Bar) error {
			for i := 1; i <= 100; i++ {
				bar.Tick(2)
				time.Sleep(1 * time.Millisecond)
				if i == 50 {
					return errors.New("connection error")
				}
			}
			return nil
		})
		pgroup.Go("Railgun Image", 50, func(c *gluey.Ctx, bar *gluey.Bar) error {
			for i := 1; i <= 100; i++ {
				bar.Tick(1)
				time.Sleep(10 * time.Millisecond)
			}
			return nil
		})
		pgroup.Wait()

		f.SetCloseTitle("Completed with failures")

		return c.InFrame("starting up env", func(c *gluey.Ctx, f *gluey.Frame) error {
			sgroup := c.NewSpinGroup()
			sgroup.Go("redis", func() error {
				time.Sleep(time.Second)
				return nil
			})
			sgroup.Go("mysql", func() error {
				time.Sleep(500 * time.Millisecond)
				return nil
			})
			sgroup.Go("elasticsearch", func() error {
				time.Sleep(2 * time.Second)
				return errors.New("elasticseach failed to start")
			})

			return c.Debreif(sgroup.Wait())
		})
	})
}
