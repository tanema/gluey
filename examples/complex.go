package main

import (
	"errors"
	"sync"
	"time"

	"github.com/tanema/gluey"
)

func main() {
	ctx := gluey.New()
	ctx.AskDefault("Username", "foo")
	ctx.Password("Password")
	ctx.Confirm("Skip Run?")
	ctx.SelectMultiple("What's your text editor", []string{"Vim", "Emacs", "Sublime", "VSCode", "Atom", "other", "Vim", "Emacs", "Sublime", "VSCode", "Atom", "other"})
	ctx.InFrame("Build", func(c *gluey.Ctx, f *gluey.Frame) error {
		f.SetShowElapsed(true)
		c.InFrame("Cloning", func(c *gluey.Ctx, f *gluey.Frame) error {
			bar := c.Progress("", 100)
			for i := 1; i <= 100; i++ {
				bar.Tick(1)
				time.Sleep(1 * time.Millisecond)
			}
			return nil
		})

		c.Println("Requesting something")
		f.Divider("Request Failed", "yellow")
		c.Println("https failed or something")

		wg := sync.WaitGroup{}
		wg.Add(3)
		pgroup := c.NewProgressGroup()
		bar1 := pgroup.Add("Git Clone", 100)
		go func() {
			defer wg.Done()
			for i := 1; i <= 100; i++ {
				bar1.Tick(1)
				time.Sleep(1 * time.Millisecond)
			}
		}()

		bar2 := pgroup.Add("Docker Image", 200)
		go func() {
			defer wg.Done()
			for i := 1; i <= 100; i++ {
				bar2.Tick(2)
				time.Sleep(1 * time.Millisecond)
				if i == 50 {
					bar2.Fail(errors.New("connection error"))
				}
			}
		}()

		bar3 := pgroup.Add("Railgun Image", 50)
		go func() {
			defer wg.Done()
			for i := 1; i <= 100; i++ {
				bar3.Tick(1)
				time.Sleep(10 * time.Millisecond)
			}
		}()
		wg.Wait()

		f.SetCloseTitle("Completed with failures")

		return c.InFrame("starting up env", func(c *gluey.Ctx, f *gluey.Frame) error {
			sgroup := c.NewSpinGroup()
			redisSpinner := sgroup.Add("redis")
			mysqlSpinner := sgroup.Add("mysql")
			esSpinner := sgroup.Add("elasticsearch")
			time.Sleep(500 * time.Millisecond)
			mysqlSpinner.Done()
			time.Sleep(time.Second)
			redisSpinner.Done()
			time.Sleep(time.Second)
			esSpinner.Fail(errors.New("elasticseach failed to start"))
			return nil
		})
	})
}
