# 🎈 [Blogo](https://blogo.site)

Blogo is a light and easy blogging engine. No complicated extras, just a straightforward blog. 

Now, here's the twist: Blogo can also publish your posts to Nostr for backing them up and getting even more reach. **Lift your ideas higher!**

## Some blogs using Blogo

- [blogo.site](https://blogo.site)
- [blog.kycnot.me](https://blog.kycnot.me)
- [blog.kyun.host](https://blog.kyun.host)

## Features

- **Easy to use**:Just put Markdown files in a folder and Blogo will take care of the rest.
- **Fast**: Blogo is written in Golang and uses BadgerDB as the backend db.
- **Light**: Blogo is light on resources, and light on your eyes:
    - No JS, no tracking, no cookies.
    - No cluttered UI, focus on reading.
    - ~10MB Docker image.
- **Markdown**: Write your posts in Markdown.
    - Github Flavored Markdown is supported.
    - Syntax Highlighting using [chroma](https://github.com/alecthomas/chroma)
    - YAML Metadata for posts info.
- **Feeds**: RSS, Atom and JSON feeds!
- **Raw endpoint**: Add `/raw` to any article link to get the raw markdown!
- **About page**: Easily create an About page so everyone can know more about you.
- **Customizable**: You can fully customize the look and feel of your blog by editing the templates and CSS.
    - Uses Golang Templates, TailwindCSS and pure plain CSS.
- **Nostr**: Publish your posts to Nostr for backing them up and getting more reach.
    - Set your own key, or let Blogo generate one for you.
    - Set your own relay list, or use the default list.
- **Auto-reload**: When a new post is added, or changed, blogo automatically reloads it.
- **SEO/SSNN Optimized** - Blogo is optimized for SEO, it contains all necessary meta tags and social sharing tags!
- **No JS**: Blogo doesn't use any JavaScript, so it's widely compatible and secure.
- **CLI Tool**: A simple CLI tool will allow you to create new post templates.

## Self-hosting using Docker Compose

The easiest way to self-host Blogo is by using Docker. 

1. Get the docker-compose.yml:

```yml
services:
  blogo:
    image: pluja/blogo:latest
    container_name: blogo
    restart: unless-stopped
    volumes:
      - ./articles:/app/articles
    ports:
      - "127.0.0.1:3000:3000"
    environment:
      # CONFIG
      BLOGO_TITLE: Blogo
      BLOGO_DESCRIPTION: A blog built with Blogo!
      BLOGO_KEYWORDS: blog,open source
      BLOGO_URL: http://localhost:3000
      #BLOGO_ANALYTICS: '<script defer src="https://my.analytics.site/script.js"></script>'
      TIMEZONE: UTC

      # NOSTR CONFIG
      PUBLISH_TO_NOSTR: false
      #NOSTR_NSEC: ""
      #NOSTR_RELAYS: "wss://nostr-pub.wellorder.net,wss://relay.damus.io,wss://relay.nostr.band"
```

2. Edit the `docker-compose.yml` file to fit your needs.

3. Run blogo:

```bash
docker compose up -d
```

Blogo is now available at [http://localhost:3000](http://localhost:3000). You can now [create your first article](#usage).

## Usage

Using Blogo is pretty simple. Once you have blogo running, you can create new articles by just running `blogo -new my-post-slug`, where `my-post-slug` is the slug of the post (used in the url). This will create a new template in the `articles` folder. Edit that file with your favorite text editor. Once done, save it and Blogo will take care of the rest (yes, it auto-reloads).

> If you're on docker, you can run `docker exec -it blogo blogo -new my-post-slug` to create a new post.

### Metadata fields

Blogo uses YAML metadata to get the post info. The metadata is located at the top of the file, between `---` and `---`.

Here's a list of the available metadata fields:

- `Title`: The title of the post. This will also be used as the title for sharing and SEO.
- `Author`: The author of the post.
- `Summary`: The summary of the post. This is used in the index page. This will also be used as the description for sharing and SEO.
- `Image`: The image of the post. This is used as the post thumbnail / header image. This will also be used as the thumbnail when sharing.
- `Tags`: The tags of the post. Must be a list of strings. This will also be used as the keywords for SEO.
- `Date`: The date of the post. Must be in the format `YYYY-MM-DD HH:MM`.
- `Draft`: Whether the post is a draft or not. Must be `true` or `false`.
- `Layout`: The layout of the post. For now, only `post` is available.
- `NostrUrl`: The url to the Nostr content. If set to `0` it will disable the posting of that article to Nostr even if Nostr publishing is enabled.

### About page

To create an about page, just create a file called `about.md` in the `articles` folder. Blogo will automatically detect it and create a link to it in the navbar.

### Static Content

To add your own static content, you can just bind-mount any folder to `/app/static/your-folder`.

For example if you are using docker compose, you can add:

```
volumes:
    - ./img:/app/static/img
```

Then you can just use `/static/img/your-image.jpg` in the markdown to add an image.

> The `/app/static` folder contains the css styles needed for styling Blogo. For this, it is recommended to always create subfolders with bind mounts inside.

### Publish to Nostr

If you set the `PUBLISH_TO_NOSTR` variable in the `docker-compose.yml` file to `true`, Blogo will publish your posts to Nostr. By default, Blogo will generate an ephemeral key (changes on every restart) and use a default relay list. 

You can change either of these defaults by setting any of these variables in the `docker-compose.yml` file:

- `NOSTR_NSEC` - expects a valid `nsec` key. If you set this key, your posts will be always published for the same key, even on restarts.
    - You can generate a new Nostr key pair using `blogo -nkeys`.
- `NOSTR_RELAY_LIST` - expects a comma-separated list of relays (with protocol); eg. `wss://relay1.com,wss://relay2.net`.

> You can avoid publishing a particular post to Nostr by setting the `NostrUrl` metadata field in the post to `false` or `0`.

> Posts are published to Nostr as [Long-Form events](https://github.com/nostr-protocol/nips/blob/master/23.md) following the definition in [NIP-33](https://github.com/nostr-protocol/nips/blob/master/33.md#referencing-and-tagging).

### Add analytics

You can add analytics to your blog by setting the `BLOGO_ANALYTICS` variable in the `docker-compose.yml` file to your analytics script. Blogo will automatically add it to the bottom of the page. **Make sure to put it all in a single line**!

```env
BLOGO_ANALYTICS='<script defer src="https://my.analytics.site/script.js"></script>'
```

## Customization

You can customize the look and feel of your blog by editing the templates and CSS. 

### Templates

The templates are located in the `templates` folder:

- `base.html`: The base template. All other templates extend this one.
    - Receives: A [Config](https://github.com/pluja/blogo/-/blob/main/blogo/models.go) struct with the name `Blogo`.
- `index.html`: The index template. This is the template used for the index page, where the posts are listed.
    - Receives: a list of articles [[]Article](https://github.com/pluja/blogo/-/blob/main/blogo/models.go) and the welcome text (string).
- `post.html`: The post template. This is the template used for the post reading page.
    - Receives: an [Article](https://github.com/pluja/blogo/-/blob/main/blogo/models.go).
- `about.html`: The about template. This is the template used for the about page.

### Styles

The templates are written in Golang Templates, and the CSS is written in TailwindCSS and pure CSS. Feel free to tweak them to your liking.

The CSS is located in the `static/css` folder. 

The main content makes use of TailwindCSS classes, so you can just tweak that to your liking. Note: You will need to rebuild the TailwindCSS using `npx` for new classes to apply.

> The rendered Markdown is styled with pure CSS. You can tweak that in the `static/css/markdown.css` file. All markdwon is wrapped inside a `div` with the `markdown` id, so you can use that to style it.

#### Adding more stylesheets

You can easily add custom stylesheets to any page. Place the stylesheet into the `static/css/ ` folder. Then, just use the `extra` block from the template to link them. (take a look at the `post.html` template and look for the `extra` block to see how it's done).
