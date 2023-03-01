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
END IF;

DELETE FROM `mqtt_history`
WHERE `id` NOT IN (
    SELECT `id` FROM (
                         SELECT `id` FROM `mqtt_history`
                         WHERE `topicid` = NEW.id
                         ORDER BY `ts` DESC
                             LIMIT 10
                     ) AS `temp_table`
) AND `topicid` = NEW.id;

END//
DELIMITER ;
SET SQL_MODE=@OLDTMP_SQL_MODE;
