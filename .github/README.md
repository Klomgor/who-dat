
<p align="center">
<img src="https://i.ibb.co/J5r1zCP/who-dat-square.png" width="128" /><br />
<i>Free & Open Source WHOIS Lookup Service</i>
<br />
<i>No-CORS, no auth API that's publicly available or easily self-hostable</i>
<br />
<b>🌐 <a href="https://who-dat.as93.net/">who-dat.as93.net</a></b><br />
</p>

---

<details>
  <summary>Contents</summary>
  
- [API Usage](#api-usage)
  - [Base URL](#base-url)
  - [Endpoints](#endpoints)
    - [Single Domain](#single-domain-lookup-domain)
    - [Bulk Domains](#multiple-domain-lookup-multi)
- [Deployment](#deployment)
  - [Option 1: Vercel](#option-1-vercel)
  - [Option 2: Docker](#option-2-docker)
  - [Option 3: Binary](#option-3-binary)
  - [Option 4: Build from Source](#option-4-build-from-source)
- [Adding Auth](#authentication)
- [Development](#development)
- [Contributing](#contributing)
- [Web Interface](#web-interface)
- [Mirror](#mirror)
- [Credits](#credits)
- [More Like This](#more-like-this)
- [License](#license)

</details>

## API Usage

> **TL;DR** Get the WHOIS records for a site: `curl https://who-dat.as93.net/example.com`

Who-Dat queries [RDAP](https://en.wikipedia.org/wiki/Registration_Data_Access_Protocol) where available and falls back to WHOIS, returning a single, consistent, normalized JSON shape for every TLD. For detailed request + response schemas, and to try the API out, you can reference the [spec](https://who-dat.as93.net/docs)

### Base URL

The base URL for the public API is [`who-dat.as93.net`](https://who-dat.as93.net)

If you're self-hosting (recommended) then replace this with your own base URL.

### Endpoints

<details>
  <summary><h4>Single Domain Lookup <code>/[domain]</code></h4></summary>

- **URL**: `/v1/whois/[domain]` (the bare `/[domain]` shorthand also works)
- **Method**: `GET`
- **Query Params**: `?raw=true` returns the original RDAP JSON or WHOIS text instead of the normalized response
- **Success Response**:
  - **Code**: 200 (the lookup succeeded, whether or not the domain is registered)
  - **Content**: Normalized WHOIS/RDAP data as JSON; the `isRegistered` field indicates registration status.
- **Error Response** (uniform envelope: `{ "error": { "code", "message", "query" } }`):
  - **400** - invalid or unparseable domain
  - **429** - rate limited (includes a `Retry-After` header)
  - **501** - no RDAP or WHOIS source for that TLD
  - **502** / **504** - upstream registry error or timeout
- **Sample Call**:

##### Command Line

```bash
curl https://who-dat.as93.net/example.com
```

##### JavaScript

```javascript
fetch('https://who-dat.as93.net/example.com')
  .then(response => response.json())
  .then(data => console.log(data))
  .catch(error => console.error('Error:', error));
```

##### Python

```python
import requests

response = requests.get('https://who-dat.as93.net/example.com')
if response.status_code == 200:
    print(response.json())
else:
    print("Error:", response.status_code)
```

</details>

<details>
  <summary><h4>Multiple Domain Lookup <code>/multi</code></h4></summary>

- **URL**: `/multi`
- **Method**: `GET`
- **Query Params**: 
  - **domains**: A comma-separated list of domains.
- **Success Response**:
  - **Code**: 200
  - **Content**: `{ "results": [ ... ] }` - an array of normalized results, one per domain.
- **Error Response**:
  - **Code**: 400 BAD REQUEST
  - **Content**: `{ "error": { "code": "INVALID_DOMAIN", "message": "missing domains parameter" } }`
- **Sample Call**:

```
curl "https://who-dat.as93.net/multi?domains=example.com,example.net"
```

</details>

[![Who-Dat API Docs](https://img.shields.io/badge/API-Docs-85EA2D?style=for-the-badge&logo=openapiinitiative&labelColor=1b2744&link=https%3A%2F%2Fwho-dat.as93.net%2Fdocs)](https://who-dat.as93.net/docs)


---

## Deployment

#### Option 1: Vercel

This is the quickest and easiest way to get up-and-running. Simply fork the repository, then login to Vercel (using GitHub), and after importing your fork, it will be deployed! There's no additional config or keys needed, and it should work just fine on the free plan.

Alternatively, just hit the button below for 1-click deploy 👇

[![1-Click Deploy to Vercel](https://img.shields.io/badge/Deploy-Vercel-ffffff?style=for-the-badge&logo=vercel&labelColor=1b2744&link=https%3A%2F%2Fwho-dat.as93.net%2Fdocs)](https://vercel.com/new/clone?repository-url=https%3A%2F%2Fgithub.com%2Flissy93%2Fwho-dat&demo-title=Who-Dat%20Demo&demo-url=https%3A%2F%2Fwho-dat.as93.net&demo-image=https%3A%2F%2Fi.ibb.co%2FJ5r1zCP%2Fwho-dat-square.png)

#### Option 2: Docker

The light-weight Docker image is published to DockerHub ([hub.docker.com/r/lissy93/who-dat](https://hub.docker.com/r/lissy93/who-dat)), as well as GHCR ([here](https://github.com/Lissy93/who-dat/pkgs/container/who-dat)).

Providing you've got Docker installed, you can get everything by running:

```shell
docker run -p 8080:8080 --dns 8.8.8.8 --dns 8.8.4.4 lissy93/who-dat
```

[![Deploy from Docker](https://img.shields.io/badge/Deploy-Docker-2496ED?style=for-the-badge&logo=docker&labelColor=1b2744&link=https%3A%2F%2Fwho-dat.as93.net%2Fdocs)](https://hub.docker.com/r/lissy93/who-dat)


#### Option 3: Binary

Head to the [Releases Tab](https://github.com/Lissy93/who-dat/releases), download and extract the pre-built executable for your system, then run it.

<details>

<summary>Example</summary>

If you're using the command line, you can do something like this<br>
Don't forget to update (v1.0) with the version number you want, and (linux-amd64) with your system's architecture.
  
```bash
# Download the binary for your system (from releases tab)
wget https://github.com/Lissy93/who-dat/releases/download/v0.9/who-dat-v0.9-linux-amd64.tar.gz -O ./who-dat.tar.gz

# Extract the compressed file
tar -xzvf who-dat.tar.gz

# Make it executable
chmod +x who-dat

# Run Who-Dat!
./who-dat
```

(Or, if you're a Microsoft fanboy, you can just double-click the `who-dat.exe` after extracting in Windows Explorer)

</details>



#### Option 4: Build from Source

Follow the setup instructions in the [Development](#development) section.<br>
Then run `go build -o who-dat ./cmd/server` to generate the binary for your system.<br>
You'll then be able to execute the newly built `./who-dat` file directly to start the application.

---

## Authentication

Authentication is optional, and can be enabled by setting the `AUTH_KEY` environment variable.

When authentication is enabled, all API requests must include the key in the Authorization header, using one of the formats indicated below.

#### Raw API Key

```
curl -H "Authorization: your-secret-key" https://who-dat.yourdomain.com/example.com
```

#### Bearer Token Format

```
curl -H "Authorization: Bearer your-secret-key" https://who-dat.yourdomain.com/example.com
```

#### X-API-Key Header

```
curl -H "X-API-Key: <your-api-key>" https://who-dat.yourdomain.com/example.com
```

If authentication is not configured (no `AUTH_KEY` set), the API will remain publicly accessible.

---

## Development

Prerequisites: You'll need [Go](https://go.dev/) installed. You will likley also want [Git](https://git-scm.com/) and/or [Docker](https://www.docker.com/). The frontend is plain HTML/Alpine.js embedded into the binary, so there's no Node build step.

```
git clone git@github.com:Lissy93/who-dat.git
cd who-dat
go mod download
```

Then run `go run ./cmd/server`

Alternativley, build the Docker container with `docker build -t who-dat .`

[![Open in GitPod](https://img.shields.io/badge/GitPod-Try_Live-FFAE33?style=for-the-badge&logo=gitpod&labelColor=1b2744&link=https%3A%2F%2Fcodeberg.org%2Falicia%2Fwho-dat)](https://gitpod.io/#https://github.com/lissy93/who-dat)
[![Open in VS Code](https://img.shields.io/badge/CodeSpaces-Try_Live-007ACC?style=for-the-badge&logo=visualstudiocode&labelColor=1b2744&link=https%3A%2F%2Fcodeberg.org%2Falicia%2Fwho-dat)](https://codespaces.new/Lissy93/who-dat)

---

## Web Interface

There's a very simple frontend included in the app. This is built with Alpine.js, so is super light-weight, and only adds about 100kb to the total executable.
The web interface is used to view WHOIS records for a given domain, and also hosts the API documentation.

<p align="center">
<img width="600" src="https://i.ibb.co/1dYcdZC/who-dat-screenshot.png" />
</p>

---

## Contributing

Contributions of any kind are welcome (and would be much appreciated!). Be sure to follow our [Code of Conduct](https://github.com/Lissy93/who-dat/blob/main/.github/CODE_OF_CONDUCT.md).<br>
If you're new to open source, I've put together some guides in [Git-In](https://github.com/Lissy93/git-into-open-source/), but feel free to reach out if you need any support.

Not a coder? You can still help, by raising bugs you find, updating docs, or consider sponsoring me on GitHub

[![Sponsor](https://img.shields.io/badge/Sponsor-Lissy93-EA4AAA?style=for-the-badge&logo=githubsponsors&labelColor=1b2744&link=https%3A%2F%2Fgithub.com%2Fsponsors%2FLissy93)](https://github.com/sponsors/Lissy93)

---

## Mirror

We've got a (non-Microsoft) mirror of this repository hosted on CodeBerg, at [codeberg.org/alicia/who-dat](https://codeberg.org/alicia/who-dat)

[![CodeBerg Mirror](https://img.shields.io/badge/Mirror-Who_Dat-2185D0?style=for-the-badge&logo=codeberg&labelColor=1b2744&link=https%3A%2F%2Fcodeberg.org%2Falicia%2Fwho-dat)](https://codeberg.org/alicia/who-dat)


---

## Credits

##### Contributors

[![contributors badge](https://readme-contribs.as93.net/contributors/lissy93/who-dat?shape=squircle)](https://github.com/lissy93/who-dat/graphs/contributors)

##### Sponsors

[![sponsors badge](https://readme-contribs.as93.net/sponsors/lissy93?shape=squircle)](https://github.com/sponsors/lissy93)

---

## License

> _**[Lissy93/Who-Dat](https://github.com/Lissy93/who-dat)** is licensed under [MIT](https://github.com/Lissy93/who-dat/blob/HEAD/LICENSE) © [Alicia Sykes](https://aliciasykes.com) 2024 - present._<br>
> <sup align="right">For information, see <a href="https://tldrlegal.com/license/mit-license">TLDR Legal > MIT</a></sup>

<details>
<summary>Expand License</summary>

```
The MIT License (MIT)
Copyright (c) Alicia Sykes <alicia@omg.com> 

Permission is hereby granted, free of charge, to any person obtaining a copy 
of this software and associated documentation files (the "Software"), to deal 
in the Software without restriction, including without limitation the rights 
to use, copy, modify, merge, publish, distribute, sub-license, and/or sell 
copies of the Software, and to permit persons to whom the Software is furnished 
to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included install 
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED,
INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANT ABILITY, FITNESS FOR A
PARTICULAR PURPOSE AND NON INFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT
HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION
OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE
SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
```

</details>


<!-- License + Copyright -->
<p  align="center">
  <i>© <a href="https://aliciasykes.com">Alicia Sykes</a> 2026</i><br>
  <i>Licensed under <a href="https://gist.github.com/Lissy93/143d2ee01ccc5c052a17">MIT</a></i><br>
  <a href="https://github.com/lissy93"><img src="https://i.ibb.co/4KtpYxb/octocat-clean-mini.png" /></a><br>
  <sup>Thanks for visiting :)</sup>
</p>

<!-- Dinosaurs are Awesome -->
<!-- 
                        . - ~ ~ ~ - .
      ..     _      .-~               ~-.
     //|     \ `..~                      `.
    || |      }  }              /       \  \
(\   \\ \~^..'                 |         }  \
 \`.-~  o      /       }       |        /    \
 (__          |       /        |       /      `.
  `- - ~ ~ -._|      /_ - ~ ~ ^|      /- _      `.
              |     /          |     /     ~-.     ~- _
              |_____|          |_____|         ~ - . _ _~_-_
-->
