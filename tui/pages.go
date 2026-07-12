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
I am a Software Engineer and researcher driven by the intersection of advanced mathematics and robust system design. While my professional roots are in building scalable server-side applications and web services (Python, Go, JavaScript), my trajectory has evolved toward solving more fundamental challenges in theoretical computer science and algebraic structures.

Currently, my primary focus lies in Abstract Algebra, Coding Theory, and Modern Cryptography. I am particularly interested in the implementation of post-quantum systems.

You can chekout my projects in "Projects" tab and my socials in "Links" tab.
`,
	}

	projects := page{
		id:    "projects",
		title: "Projects",
		desc:  "What I have been working on",
		content: `# Projects

Here are some of my projects. Not much but I wanted to mention.

## [PQC-Share](https://github.com/batt0s/pqc-share) and [McEliece RS](https://github.com/batt0s/mceliece_rs)
- I'm currently working on a cryptography project for my final year project. I'm trying to build a KEM-DEM system using Classic McEliece and AES256. You can follow the progress in GitHub.

## [Rizzy](https://github.com/batt0s/rizzy)
- Yet another interpreter made in Go

## [GoShort](https://goshort.battos.dev)
- A URL Shortener I made to learn Go
- Source code hosted on GitHub: [github.com/batt0s/goshort](https://github.com/batt0s/goshort)

You can see more on [my GitHub Profile](https://github.com/batt0s).
`,
	}

	links := page{
		id:    "links",
		title: "Links",
		desc:  "How to reach me",
		content: `# Links
- Email: [me@battos.dev](mailto:me@battos.dev) or [kerem.ullen@pm.me](mailto:kerem.ullen@pm.me)
- Website: [battos.dev](https://battos.dev) 
- Github: [@batt0s](https://github.com/batt0s)
- Blog/Pages: [pages.battos.dev](https://pages.battos.dev) 
- LinkedIn: [linkedin.com/in/kerem-ullen](https://linkedin.com/in/kerem-ullen)
- YouTube: [@packagebattos](https://www.youtube.com/@packagebattos)
`,
	}

	return []list.Item{home, about, links, projects}

}
