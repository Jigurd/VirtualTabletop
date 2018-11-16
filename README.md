# VirtualTabletop
**Project description**:

This project uses:
1. Heroku
2. OpenStack
3. Databases

Crate a better Roll20.

Users can register accounts and login.

Users can chat with eachother.


**What went well**:


**What we learned**:
1. How to serve HTML from Go. We learned two methods for doing this. As described under the *What was hard* section, we faced some challenges with one method, which lead us to using another method.

2. How to use cookies with Go.


**What was hard**:
1. As mentioned under *What we learned* we learned two methods for serving HTML. The first was using a template, by first parsing the HTML file and then executing it. This worked fine to start with, but when trying to incorporate cookies this proved otherwise. When we did this we executed the template at the start, which involves writing to the responseWriter. As it turns out, all headers and similar information must be written before any other information (as far as we understood, this is a problem with HTTP and not specific to Go). Our solution was to read the HTML file as pure text and using *func WriteString(w Writer, s string)*, and calling this at the end of the relevant handler function. This also makes it possible to add more to the HTML file without changing the actual file.


**Total hours**:
4102


# Usage
```/```

This is the index page, which doesn't hold much useful information. If logged in, it displays a welcome message.

```/register```

**GET**: Not much interesting. The page loads as one would expect.

**POST**: With the form values "username", "email", "password" and "confirm" a user can be registered to the database.

```/login```

**GET**: Not much interesitng. The page loads as one would expect.

**POST**: With the form values "username" and "password" you can log in. Redirects to "/" on successfull login.

```/chat```

Allows users to chat together.

# Heroku
This application is deployed on Heroku with the link: https://glacial-bastion-87425.herokuapp.com/
