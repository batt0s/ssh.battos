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
Hi, I'm **Kerem Üllenoğlu** (@batt0s).
I bet this was not what you expected!
But you are here. You can change pages in the list.

Software Developer, Linux Enthusiast, Math and Computer Science B.Sc. Student.
Currently interested in Coding Theory and Post-Quantum Cryptography.

> You can also connect directly with: ssh battos.dev -p 23234

Please remember this is an experimental project of mine.
There could be problems with views.
`,
	}

	about := page{
		id:    "about",
		title: "About",
		desc:  "About me!",
		content: `# About
Hi, I'm **Kerem Üllenoğlu** ([@batt0s](https://github.com/batt0s)).
I am a Software Developer and a Mathematics & Computer Science B.Sc. student based in Turkey.

I usually do server-side work (Python, Go, JavaScript), and I am currently interested in Coding Theory and Post-Quantum Cryptography. My trajectory has evolved toward solving more fundamental challenges in theoretical computer science and algebraic structures.

Currently, my primary focus lies in Abstract Algebra, Coding Theory, and Modern Cryptography. I am particularly interested in the implementation of post-quantum systems.

- Public repositories: 34
- Followers: 27
- GitHub since: April 2020

You can check out my projects in the "Projects" tab and my socials in the "Links" tab.
`,
	}

	projects := page{
		id:    "projects",
		title: "Projects",
		desc:  "What I have been working on",
		content: `# Projects

## [mceliece_rs](https://github.com/batt0s/mceliece_rs)
- A constant-time Rust implementation of the [Classic McEliece](https://classic.mceliece.org/) post-quantum cryptosystem.
- My final year project; compatible with the NIST PQC specification and Known Answer Test (KAT) vectors.
- Includes constant-time decapsulation and implicit rejection to resist timing side-channel (CCA) attacks.
- Consists of two crates: a Rust core library and Python bindings via PyO3.
- Related post: [Implementing PQC - Building Classic McEliece in Rust](https://pages.battos.dev/cryptography/2026/07/28/mceliece.html)

## [pqc-share](https://github.com/batt0s/pqc-share)
- A post-quantum safe, end-to-end encrypted peer-to-peer file sharing tool.
- Uses Classic McEliece for key encapsulation (KEM) and AES-256-GCM for data encryption.
- Transfers files between two machines over TCP.

## [rizzy](https://github.com/batt0s/rizzy)
- A simple, interpreted programming language written in Go.
- Built on top of the book *"Writing an Interpreter in Go"* to understand how interpreters work.

## [hausdorff-dist](https://github.com/batt0s/hausdorff-dist)
- A Streamlit app that interactively computes and visualizes the Hausdorff distance between compact sets.
- You can define sets as equations/inequalities and choose different metrics; critical witness vectors are plotted.
- Developed in Python (Sympy, NumPy, Matplotlib, Streamlit).

## [anime-success-predictor-ml](https://github.com/batt0s/anime-success-predictor-ml)
- A Random Forest classifier that predicts whether an anime will be a "Hit" or a "Flop" using machine learning.
- Optimized with threshold tuning for high recall.
- The dataset was collected from MyAnimeList and shared on [Kaggle](https://www.kaggle.com/datasets/battos/myanimelist-animes).

## More
You can find all my projects and current work on my GitHub profile: [github.com/batt0s](https://github.com/batt0s?tab=repositories)
`,
	}

	now := page{
		id:    "now",
		title: "Now",
		desc:  "What I'm up to right now",
		content: `# Now

*A snapshot of what I'm currently focused on.*

- Finishing my B.Sc. in Mathematics & Computer Science.
- Final year project: a constant-time Rust implementation of Classic McEliece ([mceliece_rs](https://github.com/batt0s/mceliece_rs)).
- Exploring coding theory, abstract algebra, and post-quantum cryptography.
- Writing and sharing notes on [pages.battos.dev](https://pages.battos.dev).

## Recent posts
- [Implementing PQC - Building Classic McEliece in Rust](https://pages.battos.dev/cryptography/2026/07/28/mceliece.html)
- [Her Gün Cebir - QR Kodlar ve Reed-Solomon Kodları](https://pages.battos.dev/math/2026/03/21/qr-coding.html)
- [Linux'a Giriş (Tarihçe, Distro'lar ve Kabuğa giriş)](https://pages.battos.dev/linux/2024/12/16/linux-i.html)
`,
	}

	links := page{
		id:    "links",
		title: "Links",
		desc:  "How to reach me",
		content: `# Links
- Name: Kerem Üllenoğlu (@batt0s)
- Location: Turkey
- Email: [me@battos.dev](mailto:me@battos.dev) or [kerem.ullen@pm.me](mailto:kerem.ullen@pm.me)
- Website: [battos.dev](https://battos.dev)
- Github: [@batt0s](https://github.com/batt0s)
- Blog/Pages: [pages.battos.dev](https://pages.battos.dev)
- LinkedIn: [linkedin.com/in/kerem-ullen](https://linkedin.com/in/kerem-ullen)
- YouTube: [@packagebattos](https://www.youtube.com/@packagebattos)
- SSH: ssh battos.dev -p 23234
`,
	}

	return []list.Item{home, about, now, links, projects}

}
