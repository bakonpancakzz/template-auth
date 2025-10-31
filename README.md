# 🔒 `template-auth`
Self-hosted authentication service and OAuth2 provider, built with **Go** and **React**.

Designed to act as a **central login server**, with smaller services authenticating
through the built-in **OAuth2 framework**.

## 🌟 Features
- PostgreSQL database schema ([view here](https://github.com/bakonpancakz/template-auth/blob/main/backend/include/schema.sql))
- Secure user authentication with **TOTP**, **email**, and **passwords**
- Custom user profiles (avatar, banner, accent color, etc.)
- Custom user applications with **OAuth2** support
- Internal services:
  - **Geolocation** via [IP2Location LITE](https://lite.ip2location.com/)
  - **Email templates**
    - Supported providers:
      - [AWS Simple Email Service](https://aws.amazon.com/ses/)
      - [EmailEngine](https://panca.kz/goto/emailengine)
  - **File management**
    - Supported providers:
      - [AWS S3](https://aws.amazon.com/s3/) or compatible APIs
      - Local disk storage
- Easy to set up and highly configurable
- Minimal, auditable dependency list  
- Static site for account management

> 🔰 Quickly set up a preview instance using **Docker Compose**!  
> Guide available in the [`preview`](https://github.com/bakonpancakz/template-auth/blob/main/preview) directory.

## 📷 Showcase

<p align="center">
  <img src=".github/preview_email.png" height="360"><br>
  Email prompting user to allow login from a new location
</p>