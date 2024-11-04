## Vercel

**Vercel** is a cloud platform for static sites and Serverless Functions. A lot of people use it for blogging or websites, and asteroids can be deployed there.

*PROS*:

* no need for a machine, VPS or domain name,
* can directly use an asteroid as a `http.Handler`.

*CONS*:

* slow deployment (2-3 minutes)
* requires a github account (automatically used for each deployment)
* installation can be intimidating for non-techies,
* both asteroid and themes are embedded (thus no auto-reload),
* small updates require re-deployment.

### Links

* Doc about vercel.json:  https://vercel.com/docs/projects/project-configuration
* This article helped a lot (Thanks!): https://sorcererxw.com/en/articles/vercel-go-microservice
