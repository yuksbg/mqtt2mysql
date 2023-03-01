# MQTT to MySQL data pipeline

The MQTT-to-MySQL data pipeline is a data processing solution that subscribes to all topics on an MQTT broker, receives messages published on those topics, and stores them in a MySQL database. This pipeline is useful in scenarios where you need to store data from IoT devices, sensors, or any other devices that use MQTT for communication.

The pipeline also has an optional feature where it can store historical data for each topic that's been configured to do so. Historical data is stored in a separate table and can be used for various purposes like data analysis, reporting, and machine learning.


## Requirements

* Golang 1.16 or later
* MySQL 5.7 or later
* An MQTT broker (e.g. Mosquitto)

## Installation

Clone this repository to your local machine.
Install the required dependencies using `go get`:
```
go get github.com/go-sql-driver/mysql
go get github.com/jmoiron/sqlx
go get github.com/eclipse/paho.mqtt.golang
```

Configure the database connection in the main function of `main.go`.
Configure the MQTT broker connection in the main function of main.go.

### Usage
You can run the `mqtt2mysql` application with the following command:
```go run main.go``` with following command:


The following flags are available:

| Flag              | Description                                                                                         | Default Value                    |
|-------------------|-----------------------------------------------------------------------------------------------------|----------------------------------|
| `-d`, `--db`      | MySQL DSN                                                                                           | `"user:pass@tcp(127.0.0.1:3306/app)"` |
| `-h`, `--help`    | Help for mqtt2mysql                                                                                 |                                  |
| `-p`, `--mqtt_password` | MQTT Password                                                                                  |                                  |
| `-m`, `--mqtt_server`   | MQTT DSN                                                                                          | `"tcp://127.0.0.1:1883"`         |
| `-u`, `--mqtt_username` | MQTT Username                                                                                   |                                  |

To run `mqtt2mysql` with a custom MySQL DSN and MQTT credentials, you can use the following command:

```shell
go run main.go -d "user:password@tcp(myhost:3306/mydatabase)" -m "tcp://mqtt.example.com:1883" -u "myusername" -p "mypassword"
```
If you don't provide any flags, it will use the default values.


## MySQL Schema

Before running the pipeline, you need to create the required MySQL tables. 
Here is the schema:

```sql
CREATE TABLE IF NOT EXISTS `mqtt` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `ts` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `topic` text NOT NULL,
  `value` longblob NOT NULL,
  `qos` tinyint unsigned NOT NULL,
  `retain` tinyint unsigned NOT NULL,
  `history_enable` tinyint NOT NULL DEFAULT '1',
  `history_diffonly` tinyint NOT NULL DEFAULT '1',
  PRIMARY KEY (`topic`(255)),
  UNIQUE KEY `id` (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=14800 DEFAULT CHARSET=utf8mb3;

CREATE TABLE IF NOT EXISTS `mqtt_history` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `ts` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `topicid` int unsigned NOT NULL,
  `value` longblob,
  PRIMARY KEY (`id`),
  KEY `topicid` (`topicid`),
  KEY `ts` (`ts`)
) ENGINE=InnoDB AUTO_INCREMENT=11637 DEFAULT CHARSET=utf8mb3;

SET @OLDTMP_SQL_MODE=@@SQL_MODE, SQL_MODE='ONLY_FULL_GROUP_BY,STRICT_TRANS_TABLES,NO_ZERO_IN_DATE,NO_ZERO_DATE,ERROR_FOR_DIVISION_BY_ZERO,NO_ENGINE_SUBSTITUTION';
DELIMITER //
CREATE TRIGGER `mqtt_after_insert` AFTER UPDATE ON `mqtt` FOR EACH ROW BEGIN
  IF NEW.history_enable = 1 THEN
      INSERT INTO mqtt_history SET ts=NEW.ts, topicid=NEW.id, value=NEW.value;
      DELETE FROM `mqtt_history`
      WHERE `id` NOT IN (
          SELECT `id` FROM (
                               SELECT `id` FROM `mqtt_history`
                               WHERE `topicid` = NEW.id
                               ORDER BY `ts` DESC
                                   LIMIT 100
                           ) AS `temp_table`
      ) AND `topicid` = NEW.id;
   END IF;
END//
DELIMITER ;
SET SQL_MODE=@OLDTMP_SQL_MODE;
```

To create the tables, you can run this SQL script using a MySQL client or the command line:

```shell
mysql -u <username> -p <password> <database_name> < mqtt-schema.sql
```
Replace `<username>`, `<password>`, and `<database_name>` with the appropriate values for your MySQL environment.


#### Notes regarding mqtt_history
The `mqtt_after_insert` trigger is executed after each row is updated in the `mqtt` table. When an update occurs, the trigger checks whether the `history_enable` column of the updated row is set to `1`. If it is, the trigger inserts a copy of the updated row's values (i.e. `ts`, `id`, and `value`) into the `mqtt_history` table, along with a timestamp.

The trigger also deletes any rows in `mqtt_history` that do not belong to the updated row and are not among the most recent ten rows for that `topicid`. This is achieved using a subquery that retrieves the `id` of the ten most recent rows for the updated row's `topicid`, and a `NOT IN` condition to exclude all other rows.

The history is limited to the most recent `100` messages per topic to avoid storing an excessive amount of data.



## Usage

When the pipeline is running, it subscribes to all topics on the MQTT broker and inserts incoming messages into the `mqtt` table in the configured MySQL database.
If the `history_enable` column in the `mqtt` table is set to true for a particular topic, the pipeline also inserts a copy of the message payload into the `mqtt_history` table along with a timestamp and a foreign key referencing the `mqtt` table.

The pipeline can be stopped gracefully by sending a `SIGINT` or `SIGTERM` signal to the process.

### License
This project is licensed under the MIT License. See the LICENSE file for details.
