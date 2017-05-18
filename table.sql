CREATE TABLE users (
    id MEDIUMINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(80) NOT NULL,
    pass VARCHAR(80) NOT NULL,
    token VARCHAR(80) NOT NULL
) default character set 'utf8' ENGINE=InnoDB;

CREATE TABLE access_points (
    id MEDIUMINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    ssid VARCHAR(80) NOT NULL,
    bssid VARCHAR(80) NOT NULL
) default character set 'utf8' ENGINE=InnoDB;

CREATE TABLE checkins (
    id MEDIUMINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    ts TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) default character set 'utf8' ENGINE=InnoDB;


CREATE TABLE follows (
    id MEDIUMINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    user_id MEDIUMINT NOT NULL,
    access_point_id MEDIUMINT NOT NULL,
    FOREIGN KEY(user_id) REFERENCES users(id),
    FOREIGN KEY(access_point_id) REFERENCES access_points(id)
) default character set 'utf8' ENGINE=InnoDB;
