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
	ctx.Select("What's your text editor", []string{"Vim", "Emacs", "Sublime", "VSCode", "Atom", "other"})
	ctx.InFrame("working", func(c *gluey.Ctx) error {
		c.InFrame("Cloning", func(c *gluey.Ctx) error {
			return c.Progress(100, func(c *gluey.Ctx, bar *gluey.Bar) error {
				for i := 1; i <= 100; i++ {
					bar.Tick(1)
					time.Sleep(1 * time.Millisecond)
				}
				return nil
			})
		})

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

		return c.InFrame("starting up env", func(c *gluey.Ctx) error {
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
			return sgroup.Wait()
		})
	})
}
