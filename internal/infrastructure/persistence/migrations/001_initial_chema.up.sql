CREATE TABLE `addresses` (
  `id` binary(16) NOT NULL,
  `user_id` binary(16) NOT NULL,
  `country_id` binary(16) NOT NULL,
  `city` varchar(56) NOT NULL,
  `street` varchar(255) NOT NULL,
  `house_number` varchar(15) NOT NULL,
  `apartment_number` varchar(15) DEFAULT NULL,
  `google_maps_id` varchar(255) NOT NULL,
  `formated_adress` varchar(500) NOT NULL,
  `created_at` datetime(3) NOT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `fk_addresses_country` (`country_id`),
  CONSTRAINT `fk_addresses_country` FOREIGN KEY (`country_id`) REFERENCES `countries` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE `bus_images` (
  `bus_id` binary(16) NOT NULL,
  `url` varchar(255) NOT NULL,
  KEY `fk_buses_images` (`bus_id`),
  CONSTRAINT `fk_buses_images` FOREIGN KEY (`bus_id`) REFERENCES `buses` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE `buses` (
  `id` binary(16) NOT NULL,
  `model` varchar(255) NOT NULL,
  `registration_number` varchar(8) NOT NULL,
  `year` smallint NOT NULL,
  `gps_tracker_id` varchar(255) NOT NULL,
  `lead_driver_id` binary(16) DEFAULT NULL,
  `assistant_driver_id` binary(16) DEFAULT NULL,
  `created_at` datetime(3) NOT NULL,
  `updated_at` datetime(3) NOT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `luggage_volume` int unsigned NOT NULL,
  `max_width` smallint unsigned NOT NULL,
  `max_height` smallint unsigned NOT NULL,
  `max_length` int unsigned NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uni_buses_registration_number` (`registration_number`),
  UNIQUE KEY `uni_buses_lead_driver_id` (`lead_driver_id`),
  UNIQUE KEY `uni_buses_assistant_driver_id` (`assistant_driver_id`),
  KEY `idx_buses_deleted_at` (`deleted_at`),
  CONSTRAINT `fk_buses_assistant_driver` FOREIGN KEY (`assistant_driver_id`) REFERENCES `users` (`id`),
  CONSTRAINT `fk_buses_lead_driver` FOREIGN KEY (`lead_driver_id`) REFERENCES `users` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE `connection_updates` (
  `connection_id` binary(16) NOT NULL,
  `created_at` datetime(3) NOT NULL,
  `status` enum('Registered','Canceled','Sold','Started','Finished','Stopped','Renewed','Could Not Be Finished','Departure Time Changed') NOT NULL,
  `comment` varchar(500) DEFAULT NULL,
  KEY `fk_connections_updates` (`connection_id`),
  CONSTRAINT `fk_connections_updates` FOREIGN KEY (`connection_id`) REFERENCES `connections` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE `connections` (
  `id` binary(16) NOT NULL,
  `line` smallint NOT NULL,
  `price` mediumint unsigned NOT NULL,
  `departure_country_id` binary(16) NOT NULL,
  `destination_country_id` binary(16) NOT NULL,
  `departure_time` datetime(3) NOT NULL,
  `arrival_time` datetime(3) NOT NULL,
  `google_maps_url` varchar(256) NOT NULL,
  `bus_id` binary(16) NOT NULL,
  `created_at` datetime(3) NOT NULL,
  `type` enum('Comertial','Special Asignment','Break Down Return','Break Down Replacement') NOT NULL,
  `sell_before` datetime(3) NOT NULL,
  `backpack_price` mediumint unsigned NOT NULL,
  `small_luggage_price` mediumint unsigned NOT NULL,
  `large_luggage_price` mediumint unsigned NOT NULL,
  `max_width` smallint unsigned NOT NULL,
  `max_height` smallint unsigned NOT NULL,
  `max_length` smallint unsigned NOT NULL,
  `minimal_parcel_price` mediumint unsigned NOT NULL,
  `parcel_price_per_ten_cm` mediumint unsigned NOT NULL,
  PRIMARY KEY (`id`),
  KEY `fk_connections_departure_country` (`departure_country_id`),
  KEY `fk_connections_destination_country` (`destination_country_id`),
  KEY `fk_connections_bus` (`bus_id`),
  CONSTRAINT `fk_connections_bus` FOREIGN KEY (`bus_id`) REFERENCES `buses` (`id`),
  CONSTRAINT `fk_connections_departure_country` FOREIGN KEY (`departure_country_id`) REFERENCES `countries` (`id`),
  CONSTRAINT `fk_connections_destination_country` FOREIGN KEY (`destination_country_id`) REFERENCES `countries` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE `countries` (
  `id` binary(16) NOT NULL,
  `name` varchar(50) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uni_countries_name` (`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE `email_verification_sessions` (
  `id` binary(16) NOT NULL,
  `code` char(6) NOT NULL,
  `email` varchar(255) NOT NULL,
  `expires` datetime(3) NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE `employee_availabilities` (
  `user_id` binary(16) NOT NULL,
  `status` enum('Unavailable','Sick','Other','Busy') NOT NULL,
  `date` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `comment` varchar(500) DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

DROP TABLE `logs`;

CREATE TABLE `number_verification_sessions` (
  `id` binary(16) NOT NULL,
  `code` char(6) NOT NULL,
  `number` varchar(15) NOT NULL,
  `expires` datetime(3) NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE `parcel_payments` (
  `parcel_id` binary(16) NOT NULL,
  `price` mediumint NOT NULL,
  `method` enum('Apple Pay','Card','Cash','Google Pay') NOT NULL,
  `created_at` datetime(3) NOT NULL,
  `session_id` varchar(500) NOT NULL,
  `succeeded` tinyint(1) NOT NULL,
  KEY `fk_parcels_payment` (`parcel_id`),
  CONSTRAINT `fk_parcels_payment` FOREIGN KEY (`parcel_id`) REFERENCES `parcels` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE `parcels` (
  `id` binary(16) NOT NULL,
  `user_id` binary(16) NOT NULL,
  `connection_id` binary(16) NOT NULL,
  `sender_phone_number` varchar(15) NOT NULL,
  `sender_email` varchar(255) NOT NULL,
  `reciever_phone_number` varchar(15) NOT NULL,
  `reciever_email` varchar(255) NOT NULL,
  `sender_name` varchar(255) NOT NULL,
  `sender_last_name` varchar(255) NOT NULL,
  `reciever_first_name` varchar(255) NOT NULL,
  `reciever_last_name` varchar(255) NOT NULL,
  `pick_up_adress_id` binary(16) NOT NULL,
  `drop_off_adress_id` binary(16) NOT NULL,
  `created_at` datetime(3) NOT NULL,
  `completed_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `luggage_volume` int unsigned NOT NULL,
  `width` smallint unsigned NOT NULL,
  `height` smallint unsigned NOT NULL,
  `length` smallint unsigned NOT NULL,
  `weight` smallint unsigned NOT NULL,
  `type` enum('Documents','Package') NOT NULL,
  `qr_code` blob NOT NULL,
  PRIMARY KEY (`id`),
  KEY `fk_parcels_drop_off_adress` (`drop_off_adress_id`),
  KEY `fk_parcels_pick_up_adress` (`pick_up_adress_id`),
  CONSTRAINT `fk_parcels_drop_off_adress` FOREIGN KEY (`drop_off_adress_id`) REFERENCES `addresses` (`id`),
  CONSTRAINT `fk_parcels_pick_up_adress` FOREIGN KEY (`pick_up_adress_id`) REFERENCES `addresses` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE `passengers` (
  `id` binary(16) NOT NULL,
  `ticket_id` binary(16) NOT NULL,
  `first_name` varchar(255) NOT NULL,
  `last_name` varchar(255) NOT NULL,
  `created_at` datetime(3) NOT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE `rows` (
  `id` binary(16) NOT NULL,
  `bus_id` binary(16) DEFAULT NULL,
  `number` tinyint NOT NULL,
  PRIMARY KEY (`id`),
  KEY `fk_buses_structure` (`bus_id`),
  CONSTRAINT `fk_buses_structure` FOREIGN KEY (`bus_id`) REFERENCES `buses` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE `seat_positions` (
  `row_id` binary(16) DEFAULT NULL,
  `seat_number` tinyint NOT NULL,
  `type` enum('Space','Table','Seat') NOT NULL,
  `position` tinyint NOT NULL,
  KEY `fk_rows_positions` (`row_id`),
  CONSTRAINT `fk_rows_positions` FOREIGN KEY (`row_id`) REFERENCES `rows` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE `seats` (
  `id` binary(16) NOT NULL,
  `bus_id` binary(16) DEFAULT NULL,
  `number` tinyint NOT NULL,
  `type` enum('Window','Single','Single-Window','Aisle','Middle') NOT NULL,
  `direction` enum('Forward','Backward') NOT NULL,
  PRIMARY KEY (`id`),
  KEY `fk_buses_seats` (`bus_id`),
  CONSTRAINT `fk_buses_seats` FOREIGN KEY (`bus_id`) REFERENCES `buses` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE `stop_updates` (
  `stop_id` binary(16) NOT NULL,
  `status` enum('Confirmed','Missed','Completed') DEFAULT NULL,
  `comment` varchar(500) DEFAULT NULL,
  `created_at` datetime(3) NOT NULL,
  KEY `fk_stops_updates` (`stop_id`),
  CONSTRAINT `fk_stops_updates` FOREIGN KEY (`stop_id`) REFERENCES `stops` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE `stops` (
  `id` binary(16) NOT NULL,
  `ticket_id` binary(16) DEFAULT NULL,
  `parcel_id` binary(16) DEFAULT NULL,
  `connection_id` binary(16) NOT NULL,
  `type` enum('Passenger','Parcel') DEFAULT NULL,
  `location_type` enum('Pick-up','Drop-off') DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `fk_stops_ticket` (`ticket_id`),
  KEY `fk_stops_parcel` (`parcel_id`),
  KEY `fk_connections_stops` (`connection_id`),
  CONSTRAINT `fk_connections_stops` FOREIGN KEY (`connection_id`) REFERENCES `connections` (`id`),
  CONSTRAINT `fk_stops_parcel` FOREIGN KEY (`parcel_id`) REFERENCES `parcels` (`id`),
  CONSTRAINT `fk_stops_ticket` FOREIGN KEY (`ticket_id`) REFERENCES `tickets` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE `ticket_payments` (
  `ticket_id` binary(16) NOT NULL,
  `price` mediumint NOT NULL,
  `method` enum('Apple Pay','Card','Cash','Google Pay') NOT NULL,
  `created_at` datetime(3) NOT NULL,
  `session_id` varchar(500) NOT NULL,
  `succeeded` tinyint(1) NOT NULL,
  PRIMARY KEY (`ticket_id`),
  CONSTRAINT `fk_tickets_payment` FOREIGN KEY (`ticket_id`) REFERENCES `tickets` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE `ticket_seats` (
  `ticket_id` binary(16) NOT NULL,
  `seat_id` binary(16) NOT NULL,
  KEY `fk_ticket_seats_seat` (`seat_id`),
  KEY `fk_tickets_seats` (`ticket_id`),
  CONSTRAINT `fk_ticket_seats_seat` FOREIGN KEY (`seat_id`) REFERENCES `seats` (`id`),
  CONSTRAINT `fk_tickets_seats` FOREIGN KEY (`ticket_id`) REFERENCES `tickets` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE `tickets` (
  `id` binary(16) NOT NULL,
  `user_id` binary(16) NOT NULL,
  `connection_id` binary(16) NOT NULL,
  `phone_number` varchar(15) NOT NULL,
  `email` varchar(255) NOT NULL,
  `pick_up_adress_id` binary(16) NOT NULL,
  `drop_off_adress_id` binary(16) NOT NULL,
  `created_at` datetime(3) NOT NULL,
  `completed_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `luggage_volume` mediumint unsigned NOT NULL,
  `qr_code` blob NOT NULL,
  PRIMARY KEY (`id`),
  KEY `fk_tickets_pick_up_adress` (`pick_up_adress_id`),
  KEY `fk_tickets_drop_off_adress` (`drop_off_adress_id`),
  CONSTRAINT `fk_tickets_drop_off_adress` FOREIGN KEY (`drop_off_adress_id`) REFERENCES `addresses` (`id`),
  CONSTRAINT `fk_tickets_pick_up_adress` FOREIGN KEY (`pick_up_adress_id`) REFERENCES `addresses` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE `trip_updates` (
  `trip_id` binary(16) NOT NULL,
  `status` enum('Registered','Canceled','Changed Bus','Started','Outbound Done','Break Down','Broken Bus Fixed','Broken Bus Replaced','Finished') NOT NULL,
  `created_at` datetime(3) NOT NULL,
  `comment` varchar(500) DEFAULT NULL,
  KEY `fk_trips_updates` (`trip_id`),
  CONSTRAINT `fk_trips_updates` FOREIGN KEY (`trip_id`) REFERENCES `trips` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE `trips` (
  `id` binary(16) NOT NULL,
  `outbound_connection_id` binary(16) NOT NULL,
  `return_connection_id` binary(16) NOT NULL,
  PRIMARY KEY (`id`),
  KEY `fk_trips_outbound_connection` (`outbound_connection_id`),
  KEY `fk_trips_return_connection` (`return_connection_id`),
  CONSTRAINT `fk_trips_outbound_connection` FOREIGN KEY (`outbound_connection_id`) REFERENCES `connections` (`id`),
  CONSTRAINT `fk_trips_return_connection` FOREIGN KEY (`return_connection_id`) REFERENCES `connections` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;


CREATE TABLE `users` (
  `id` binary(16) NOT NULL,
  `first_name` varchar(50) NOT NULL,
  `last_name` varchar(50) NOT NULL,
  `date_of_birth` date NOT NULL,
  `phone_number` varchar(15) DEFAULT NULL,
  `email` varchar(255) NOT NULL,
  `password` varchar(255) NOT NULL,
  `image_url` varchar(255) NOT NULL,
  `role` enum('Customer','Admin','Driver','Support') NOT NULL,
  `created_at` datetime(3) NOT NULL,
  `updated_at` datetime(3) NOT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uni_users_email` (`email`),
  KEY `idx_users_email` (`email`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
