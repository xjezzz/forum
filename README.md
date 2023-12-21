# forum-advanced-features optional project

## Moderation features added

Added 4 types of users :

### Guests
    
    These are unregistered-users that can neither post, comment, like or dislike a post.
    They only have the permission to see those posts, comments, likes or dislikes.

### Users

    These are the users that will be able to create, comment, like or dislike posts.

### Moderators

    Moderators are users that have a granted access to special functions :
        They should be able to monitor the content in the forum by deleting or reporting post to the admin
        To create a moderator the user should request an admin for that role

### Administrators

    Users that manage the technical details required for running the forum. This user must be able to :
        Promote or demote a normal user to, or from a moderator user.
        Receive reports from moderators. If the admin receives a report from a moderator, he can respond to that report
        Delete posts and comments
        Manage the categories, by being able to create and delete them.


## Security feautures added

Add https, rate limiting, password hashing

## Image upload added

Now registered users have the possibility to create a post containing an image as well as text.
Supported extensions: JPEG, PNG and GIF types
The max size of the images to load = 20 mb

## Oauth2 added

Now you can autorize with Google or GitHub account

## Instruction
Copy the code below to the console

 ```
 git clone git@git.01.alem.school:ymukhame/forum-moderation.git
 ```

For run program
```
make run
```
and click on link (https://0.0.0.0:8080)

For built
```
make build
```

For run with docker

```
make docker-run
```
and click on link (https://0.0.0.0:8080)

For delete images & containers

```
make docker-delete
```

### Project tasks & description

This project consists in creating a web forum that allows :

- communication between users.
- associating categories to posts.
- liking and disliking posts and comments.
- filtering posts.

#### SQLite

In order to store the data in your forum (like users, posts, comments, etc.) you will use the database library SQLite.

- You must use at least one SELECT, one CREATE and one INSERT queries.

#### Authentication

In this segment the client must be able to `register` as a new user on the forum, by inputting their credentials. You also have to create a `login session` to access the forum and be able to add posts and comments.

You should use cookies to allow each user to have only one opened session. Each of this sessions must contain an expiration date. It is up to you to decide how long the cookie stays "alive". The use of UUID is a Bonus task.

Instructions for user registration:

- Must ask for email
    - When the email is already taken return an error response.
- Must ask for username
- Must ask for password
    - The password must be encrypted when stored (this is a Bonus task)

The forum must be able to check if the email provided is present in the database and if all credentials are correct. It will check if the password is the same with the one provided and, if the password is not the same, it will return an error response.

#### Communication

In order for users to communicate between each other, they will have to be able to create posts and comments.

- Only registered users will be able to create posts and comments.
- When registered users are creating a post they can associate one or more categories to it.
    - The implementation and choice of the categories is up to you.
- The posts and comments should be visible to all users (registered or not).
- Non-registered users will only be able to see posts and comments.

#### Likes and Dislikes

Only registered users will be able to like or dislike posts and comments.

The number of likes and dislikes should be visible by all users (registered or not).

#### Filter

You need to implement a filter mechanism, that will allow users to filter the displayed posts by :

- categories
- created posts
- liked posts

You can look at filtering by categories as subforums. A subforum is a section of an online forum dedicated to a specific topic.

Note that the last two are only available for registered users and must refer to the logged in user.

#### Docker

For the forum project you must use Docker.

### Used packages

- All [standard Go](https://golang.org/pkg/) packages are allowed.
- [sqlite3](https://github.com/mattn/go-sqlite3)
- [UUID](https://github.com/gofrs/uuid)
- [BCRYPT](https://cs.opensource.google/go/x/crypto/bcrypt)

# Authors
- @ymukhame
