# VirtualTabletop
## Project description

Create a better Roll20.

Users can register accounts and login.

Users can chat with eachother.


**What went well**:


**What we learned**:
1. How to serve HTML from Go. We learned two methods for doing this. As described in point 1 under *What was hard*, we faced some challenges with one method, which lead us to using another method.

2. How to use cookies with Go.

3. How to use OpenStack and deploy apps there.

4. The Linux command ```screen```. This is very useful for OpenStack deployment.


**What was hard**:
1. As mentioned under *What we learned* we learned two methods for serving HTML. The first was using a template, by first parsing the HTML file and then executing it. This worked fine to start with, but when trying to incorporate cookies this proved otherwise. When we did this we executed the template at the start, which involves writing to the responseWriter. As it turns out, all headers and similar information must be written before any other information (as far as we understood, this is a problem with HTTP and not specific to Go). Our solution was to read the HTML file as pure text and using *func WriteString(w Writer, s string)*, and calling this at the end of the relevant handler function. This also makes it possible, and fairly easy, to add more to the HTML output without changing the actual file.


**Total hours**:
4102 (sike)

**This project uses**:
1. Heroku
2. OpenStack
3. Databases (MongoDB)

# Usage
**/**

This is the index page, which doesn't hold much useful information. If logged in, it displays a welcome message.

**/register**

```POST```: With the form values "username", "email", "password" and "confirm" a user can be registered to the database.

**/login**

```POST```: With the form values "username" and "password" you can log in. Redirects to "/" on successfull login.


**/profile**



**/chat**

Allows users to chat together.


**/api/count**

```GET```: Returns a JSON with how many users are registered in the database.

Response body:


```
{

    "count": <count>
    
}
```

# Dice Parser Syntax

Our service is meant to facilitate playing tabletop roleplaying games, so emulating dicerolls is crucial. The syntax is quite simple - in the chat, you just put the desired roll in square brackets [] like so:

```
Billy Barbarian rolls to attack: [d20+9] vs AC, and does [2d6+8] damage.
```


Anything to the left of the "d" will be interpreted as adding to the amount of dice rolled ([3+2d6] rolls 5d6) and anything to the right will be interpreted as a flat modifier ([d4+8] will be somewhere between 9 and 12).

This is very handy for some systems, and means that if you want to roll multiple kinds of dice in one query, you do that like so:

```
Rick the Rogue rolls for damage: [d8+[2d6]+6]
```

The parser supports common mathematical operators (+-*/), which is also sometimes useful.

```
The party defeated the eight goblins, and get [8*75] experience points
```


# Heroku
This application is deployed on Heroku with the link: https://glacial-bastion-87425.herokuapp.com/


# Clock trigger
An independent application sends a GET request every 10 minutes to */api/count* and if the count has changed since the last check it notifies a Discord channel with how many users there are. This application is deployed on OpenStack, and the source code resides
in this repo in the folder ```clocktrigger```.


# Other notes
1. The *HTML* folder is present two places, at the top level and in the *web* folder. This is not an accidental duplication. Since our ```main.go``` file is at the top level, this is where the program looks for ```html/<htmlfile.html>``` when referenced in the code. ```handler.go``` lies in the ```web``` folder, and when the handler functions are ran through the tests, they will look for ```web/html/<htmlfile.html>```. This causes an error when they normally can't be found, making the tests fail. The tests obviously don't need the HTML itself, but when the handlers can't find the files they abort. Having the HTML in two locations is a hacky solution, and not a very good one, but we feel it is better than changing the actual code to pass the test cases (by for example looking for the files in two locations). 
