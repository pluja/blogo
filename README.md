# 🎈 [Blogo](https://blogo.site)

As easy as blogging can get. Set it up in three steps, get a handful of useful features and a nice blog.

No complicated extras, just a straightforward blog.

## Features

- **Easy to use**: Just put Markdown files in a folder and Blogo will take care of the rest.
- **Light**: Blogo is light on resources, and light on your eyes:
    - No JS, no cookies.
    - Zero-clutter UI, focus on what matters.
    - Tiny ~10MB Docker image.
    - No database, [just files](https://stephango.com/file-over-app).
    - Fast cache for popular posts.
- **Markdown**: Write your posts in Markdown.
    - Github Flavored Markdown is supported.
    - Syntax Highlighting using [chroma](https://github.com/alecthomas/chroma)
    - YAML Metadata for posts info.
- **Feeds**: RSS, Atom and JSON feeds!
- **Raw endpoint**: Add `/raw` to any article link to get the raw markdown!
- **About page**: Easily create an About page for your blog. Just name a file `about.md` and that's it.
- **Themes**: You can create your own themes, or use one of the 23 available themes.
- **Nostr**: Optionally publish your posts to Nostr for backing them up and getting more reach.
    - Set your own key, or let Blogo generate one for you.
    - Set your own relay list, or use the default list.
- **Auto-reload**: When a new post is added, or changed, blogo automatically reloads it.
- **SEO/SSNN Optimized**: Blogo is optimized for SEO, it contains all necessary meta tags and social sharing tags!
- **No JS**: Blogo doesn't use any JavaScript, so it's widely compatible and secure.
- **CLI Tool**: A simple CLI tool will allow you to create new post templates.

## Self-hosting using Docker Compose

The easiest way to self-host Blogo is by using Docker. 

1. Create the docker-compose.yml:

```yml
services:
  blogo:
    image: pluja/blogo:v3
    container_name: blogo
    restart: unless-stopped
    volumes:
      - ./articles:/blogo/articles
      - ./blogo.yml:/blogo/blogo.yml
    ports:
      - "127.0.0.1:1337:1337"
```

2. Get and edit the [config file]():

```
wget -O blogo.yml https://raw.githubusercontent.com/pluja/blogo/refs/heads/v3/example.blogo.yml
```

> All blogo.yml variables can be set as environment variables. Check out [Configuration](#configuration) section.

3. Run blogo:

```bash
docker compose up -d
```

Blogo is now available at [http://localhost:1337](http://localhost:1337). You can now [create your first article](#usage).

## Configuration

You can either configure blogo using the [`blogo.yml`](#configuration) file or env variables. The config file will be expected to be in either the same path as the exectuable (`./blogo.yml`), in `/blogo/blogo.yml` or in `$HOME/.blogo/blogo.yml`.

| Config              | Env Variable              | Default                                                                             | Description                     |
| ------------------- | ------------------------- | ----------------------------------------------------------------------------------- | ------------------------------- |
| `title`             | `BLOGO_TITLE`             | "Blogo"                                                                             | The title of the blog           |
| `description`       | `BLOGO_DESCRIPTION`       | "Welcome to my blogo"                                                               | A brief description of the blog |
| `keywords`          | `BLOGO_KEYWORDS`          | -                                                                                   | Keywords for SEO                |
| `host`          | `BLOGO_host`          | -                                                                                   | The base URL of the blog        |
| `timezone`          | `BLOGO_TIMEZONE`          | "UTC"                                                                               | The timezone for the blog       |
| `analytics`         | `BLOGO_ANALYTICS`         | -                                                                                   | Analytics script                |
| `theme`             | `BLOGO_THEME`             | "blogo"                                                                             | The theme of the blog           |
| `upvotes`           | `BLOGO_UPVOTES`           | -                                                                                   | Enable upvotes feature          |
| `articles.path`     | `BLOGO_ARTICLES_PATH`     | -                                                                                   | Path to articles directory      |
| `powered_by_footer` | `BLOGO_POWERED_BY_FOOTER` | true                                                                                | Show "Powered by" footer        |
| `nostr.publish`     | `BLOGO_NOSTR_PUBLISH`     | false                                                                               | Enable Nostr publishing         |
| `nostr.nsec`        | `BLOGO_NOSTR_NSEC`        | -                                                                                   | Nostr secret key                |
| `nostr.relays`      | `BLOGO_NOSTR_RELAYS`      | ["wss://nostr-pub.wellorder.net", "wss://relay.damus.io", "wss://relay.nostr.band"] | Nostr relay servers             |


## Usage

Using Blogo is pretty simple. Once you have blogo running, you can create new articles by just running `blogo new my-post-slug`, where `my-post-slug` is the slug of the post (used in the url). This will create a new template in the `articles` folder. Edit that file with your favorite text editor. Once done, save it and Blogo will take care of the rest (yes, it auto-reloads).

> If you're on docker, you can run `docker exec -it blogo blogo new my-post-slug` to create a new post.

### Article Metadata fields

Articles in Blogo use YAML metadata. The metadata is located at the top of the file, between `---` and `---`.

Here's a list of the available metadata fields:

- `Title`: The title of the post. This will also be used as the title for sharing and SEO.
- `Author`: The author of the post.
- `Summary`: The summary of the post. This is used in the index page. This will also be used as the description for sharing and SEO.
- `Image`: The image of the post. This is used as the post thumbnail / header image. This will also be used as the thumbnail when sharing.
- `Tags`: The tags of the post. Must be a list of strings. This will also be used as the keywords for SEO.
- `Date`: The date of the post. Must be in the format `YYYY-MM-DD HH:MM`.
- `Draft`: Whether the post is a draft or not. Must be `true` or `false`.
- `Layout`: The layout of the post. For now, only `post` is available.
- `Nostr`: If set to a falsey value (`false`, `0`...) it will disable the posting of that article to Nostr even if Nostr publishing is enabled.

Example metadata, it must be at the top of the `.md` file:

```yaml
---
Author: John Doe
Date: 2023-04-28 17:54
Draft: false
Image: https://picsum.photos/1920/1080
Summary: In the grand tapestry of human history, few periods have been as transformative or as rapidly evolving as the present.
Tags:
    - some
    - tags
Title: Example Post
---
```

### About page

To create an about page, just create a file called `about.md` in the `articles` folder. Blogo will automatically detect it and create a link to it in the navbar.

### Static Content

To add your own static content, you can just bind-mount any folder to `/app/static/your-folder`.

For example if you are using docker compose, you can add:

```
volumes:
    - ./img:/blogo/frontend/static/img
```

Then you can just use `/static/img/your-image.jpg` in the markdown to add an image:

```md
![My image](/static/img/your-image.jpg)
```

> The `/app/static` folder contains the css styles needed for styling Blogo. For this, it is recommended to always create subfolders with bind mounts inside.

### Publish to Nostr

If you set the `nostr.publish` variable in the [`blogo.yml`](#configuration) file to `true`, Blogo will publish your posts to Nostr. By default, Blogo will generate an ephemeral key (changes on every restart) and use a default relay list. 

You can change either of these defaults by setting any of these variables in the [`blogo.yml`](#configuration) file:

- `nostr.nsec` - expects a valid `nsec` key. If you set this key, your posts will be always published for the same key, even on restarts.
    - You can generate a new Nostr key pair using `blogo -nkeys`.
- `nostr.relays` - expects a yaml list of relays (with protocol); eg. `["wss://relay1.com","wss://relay2.net"]`.

> You can avoid publishing a particular post to Nostr by setting the `Nostr` metadata field in the post to `false` or `0`.

> Posts are published to Nostr as [Long-Form events](https://github.com/nostr-protocol/nips/blob/master/23.md) following the definition in [NIP-33](https://github.com/nostr-protocol/nips/blob/master/33.md#referencing-and-tagging). 

### Add analytics

You can add analytics to your blog by setting the `analytics` variable in the [`blogo.yml`](#configuration) file to your analytics script. Blogo will automatically add it to the bottom of the page. **Make sure to put it all in a single line**!

```yaml
analytics: '<script defer src="https://my.analytics.site/script.js"></script>'
```

## Customization

You can customize the look and feel of your blog by editing the templates and CSS. 

### Themes

The easiest way to customize your Blogo, is to create a new theme. For this, you can add a new `.css` file to the `static/css/themes/` folder with the contents:

```css
:root {
    --blogo-background: rgb(26, 26, 28);
    --blogo-primary: #4d79bf;
    --blogo-primary-emphasis: #567d98;
    --blogo-primary-dark: #141626;
    --blogo-primary-light: #bbbdc4;
    --blogo-text: #c8c7c1;
    --blogo-text-select: #353839;
    --blogo-text-fade: rgb(81, 81, 87);
    --blogo-text-title: rgb(206, 205, 195);
}
```

And customize the colors from there. If you're in docker, you can create a volume like:

```yaml
volumes:
    - /path/to/mytheme.css:/blogo/static/css/themes/mytheme.css
```

Then set `mytheme` as the value in `theme` on your [`blogo.yml`](#configuration) config file.

## Change favicon and banner

Simply replace the `frontend/static/assets/favicon.webp` and `frontend/static/assets/banner.webp` files. They must be named as `favicon.webp` and `banner.webp`. If you're on Docker, you can bind files like:

```yaml
volumes:
    - ./path/to/my/favicon.web:/blogo/frontend/static/assets/favicon.webp
    - ./path/to/my/banner.web:/blogo/frontend/static/assets/banner.webp
```

### Templates

The templates are located in the `src/frontend/templates` folder.

## Some blogs using Blogo

- [blogo.site](https://blogo.site)
- [blog.kycnot.me](https://blog.kycnot.me)
- [blog.kyun.host](https://blog.kyun.host)
