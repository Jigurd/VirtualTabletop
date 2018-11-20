# VirtualTabletop
## Group members
- Sigurd Stapnes (sigurdzs)
- Håkon Schia (haakosc)
- Benjamin Bergseth (benjabe)
- Victor Sebastian Standal Clausen (vsclause)
- Jon Gunnar Fossum (jongfo)
- Leon Nicola Cinquemani (leonnc)

## Project description

Virtual Tabletop service: fixing a lot of the small problems with services like Roll20 (sub-par dice parser, bad token system, clunky page navigation) and adding a few killer features, specifically the ability to import and export character sheets with JSON (technically possible with Roll20, but it's a massive pain in the neck). If we're feeling particularly ambitious, we might even make it possible for Game Masters to share monster statblocks in a searchable database.

Users can register accounts.

After they have registered they can:
0. Log in
1. Chat with eachother.
2. Join games
3. Create and edit characters
4. Edit user information
5. Log out


**What went well**:


**What we learned**:
1. How to serve HTML from Go. We learned two methods for doing this: Templates and reading the file as is and using ```func WriteString(w Writer, s string)```. For static HTML files (or very basic editing), using the second method is probably the easiest and most straight forward, but using templates allows for a lot more dynamic websites. The syntax isn't exactly the prettiest, though.

2. How to use cookies with Go (which wasn't as straight forward as one might think, see point 1 under *What was hard*)

3. How to use OpenStack and deploy apps there.

4. The Linux command ```screen```. This is very useful for OpenStack deployment since you can close the terminal without terminating the program.


**What was hard**:
1. When figuring out how to use cookies we had a lot of problems. At the start we were parsing the html to a template and executing it at the top of the handlers. This was not a very good idea. As it turns out, headers and cookies need to be set before anything is written to the responseWriter (as far as we understand this is a "problem" with HTTP and not specific to Go), so when we tried to set the cookies, nothing was saved. This was obviously a fairly easy fix. All that is needed is to execute or write the HTML at the end, but nevertheless this was time consuming to figure out.
2. Uploading images by letting the user browse their computer and uploading local files. We spent a decent amount of time on this, but didn't manage it at all.

**Total hours**:
94.25

**This project uses**:
1. Heroku
2. OpenStack
3. Databases (MongoDB)

# Usage
**/**

This is the index page. If logged in, it displays a welcome message and what games the user is a part of.

**/register**

```POST```: With the form keys "username", "email", "password" and "confirm" (confirm password) a user can be registered to the database.

**/login**:

```POST```: With the form keys "username" and "password" you can log in. Redirects to "/profile" on successfull login.


**/logout**:

Logs a user out.


**/playerdirectory**:

Shows a list of users available to play (being shown here can be toggled, see more below), along with their user description.


**/profile**:

```POST```: With the form keys "visible" (value "visible" or "notvisible", this updates a users visibility under ```/playerdirectory```) and "desc" you can update user information (only if you are looged in).

When not logged in a link to /register and /login is shown.


**/game/{id}**:

Shows information about a given game.


**/i/{id}**:

Invite links; if you are logged in while visiting a valid ID you join the corresponding game, if there is space for more players.


**/u/{username}**:

Shows information about the user.


**/newgame**:

Allows users to create new games


**/chat**:

Allows users to chat together.


**/api/usercount**:

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
