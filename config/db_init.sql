CREATE TABLE IF NOT EXISTS users(
		id         INTEGER PRIMARY KEY AUTOINCREMENT,
		email      TEXT      NOT NULL unique,
		pass_hash  TEXT      NOT NULL,
		nickname   TEXT      NOT NULL unique,
		roles      TEXT 	 NOT NULL,
		request   BOOLEAN 	 NOT NULL
		);
CREATE TABLE IF NOT EXISTS tags(
		id         INTEGER PRIMARY KEY AUTOINCREMENT,
		name       TEXT      NOT NULL unique
	);
CREATE TABLE IF NOT EXISTS posts(
		id         INTEGER 	  PRIMARY KEY AUTOINCREMENT,
		user_id    INTEGER    NOT NULL,
		title      TEXT       NOT NULL,
		body       TEXT       NOT NULL,
		img_name   TEXT,
		FOREIGN KEY (user_id) REFERENCES users (id)
	);
CREATE TABLE IF NOT EXISTS posts_tags(
		post_id    INTEGER        NOT NULL,
		tag_id     INTEGER        NOT NULL,
		FOREIGN KEY (post_id) REFERENCES posts (id),
		FOREIGN KEY (tag_id) REFERENCES tags (id)
	);
CREATE TABLE IF NOT EXISTS comments(
		id         INTEGER     PRIMARY KEY AUTOINCREMENT,
		post_id    INTEGER     NOT NULL,
		user_id    INTEGER     NOT NULL,
		body       TEXT        NOT NULL,
		FOREIGN KEY (user_id) REFERENCES users (id),
		FOREIGN KEY (post_id) REFERENCES posts (id)
	);
CREATE TABLE IF NOT EXISTS reactions(
		id         INTEGER 		 PRIMARY KEY AUTOINCREMENT,
		is_like    BOOLEAN   	 NULL,
		user_id    INTEGER       NOT NULL,
		comment_id INTEGER,
		post_id    INTEGER,
		FOREIGN KEY (user_id) 	 REFERENCES users (id),
		FOREIGN KEY (comment_id) REFERENCES comments (id),
		FOREIGN KEY (post_id)	 REFERENCES posts (id)
 	);
CREATE TABLE IF NOT EXISTS sessions(
		user_id    INTEGER       NOT NULL,
		token      TEXT       	 NOT NULL,
		expired    TIMESTAMP     NOT NULL,
		FOREIGN KEY (user_id)	 REFERENCES users (id)
 	);
CREATE TABLE IF NOT EXISTS reports(
		id         INTEGER PRIMARY KEY AUTOINCREMENT,
		nickname   TEXT			  NOT NULL,
		post_id    INTEGER        NOT NULL,
		report     TEXT			  NOT NULL,
		status 	   BOOLEAN,
		FOREIGN KEY (post_id) REFERENCES posts (id)
 	);

CREATE TABLE IF NOT EXISTS actions(
		id 			INTEGER 		PRIMARY KEY AUTOINCREMENT,
		post_id 	INTEGER 		NOT NULL,
		user_id		INTEGER 		NOT NULL,
		by_user_id  INTEGER 		NOT NULL,
		action 		TEXT			NOT NULL,
		is_read		BOOLEAN			NOT NULL,
		FOREIGN KEY (post_id) 		REFERENCES posts (id),
		FOREIGN KEY (by_user_id) 	REFERENCES users (id),
		FOREIGN KEY (user_id) 		REFERENCES users (id)

)