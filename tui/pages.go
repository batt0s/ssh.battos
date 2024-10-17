package tui

import (
	"github.com/charmbracelet/bubbles/list"
)

func getSelectedPage(pageId string) page {
	pages := Pages()
	for _, p := range pages {
		p, _ := p.(page)
		if p.id == pageId {
			return p
		}
	}
	return page{}
}

type page struct {
	id      string
	title   string
	desc    string
	content string
}

func (p page) Title() string {
	return p.title
}

func (p page) Description() string {
	return p.desc
}

func (p page) FilterValue() string {
	return p.title
}

func Pages() []list.Item {
	home := page{
		id:    "home",
		title: "Home",
		desc:  "Welcome!",
		content: `# Welcome!
I bet this was not what you expected!
But you are here. You can change pages in the list.

Please remember this is a experimental project of mine.
There could be problems with views.
`,
	}

	about := page{
		id:    "about",
		title: "About",
		desc:  "About me!",
		content: `# About
Hello, I'm Kerem, a software developer with experience in Python, specializing in server-side application development.
I'm also familiar with JavaScript and Go.

In my professional journey, I've actively participated in projects focusing on creating web services for artificial intelligence based applications.

As a dedicated Linux enthusiast, I have a passion for experimenting with new programming languages and try open-source projects.
Additionally, I find joy in exploring the mathematical foundations behind the functionality of software.

My commitment to continuous self-improvement in the software world drives me to seek new and challenging projects, aiming to expand my skill set and gain valuable experience.

You can chekout my projects in "Projects" tab and my socials in "Links" tab.
`,
	}

	projects := page{
		id:    "projects",
		title: "Projects",
		desc:  "What I have been working on",
		content: `# Projects

Here are some of my projects. Not much but I wanted to mention.

## [Rizzy](https://github.com/batt0s/rizzy)
- Yet another interpreter made in Go

### What I have learned?
- Basics of designing an interpreted programming language
- Implementing lexical analysis and parsing in Go
- Handling different data types and operations in a language
- Iterative improvements based on a foundational interpreter architecture

## [GoShort](https://goshort.battos.dev)
- A URL Shortener I made to learn Go
- Source code hosted on GitHub: [github.com/batt0s/goshort](https://github.com/batt0s/goshort)

### What I have learned?
- Using Go with net/http and [chi](https://github.com/go-chi/chi) to make RestAPIs
- Using PostgreSQL with Go
- Writing Unit Tests for Go APIs
- Consuming an API with Javascript
- Using GitHub actions

`,
	}

	links := page{
		id:    "links",
		title: "Links",
		desc:  "How to reach me",
		content: `# Links
- Email: [kerem.ullen@pm.me](mailto:kerem.ullen@pm.me)
- Website: [battos.dev](https://battos.dev) 
- Github: [@batt0s](https://github.com/batt0s)
- Blog/Pages: [pages.battos.dev](https://pages.battos.dev) 
- LinkedIn: [linkedin.com/in/kerem-ullen](https://linkedin.com/in/kerem-ullen)
- YouTube: [@packagebattos](https://www.youtube.com/@packagebattos)
`,
	}

	return []list.Item{home, about, links, projects}

}
