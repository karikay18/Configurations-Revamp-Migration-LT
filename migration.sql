--  Creating a new table for configurations

CREATE TABLE `configurations` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `organization_id` int NOT NULL,
  `name` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL,
  `platform` enum('real-device-mobile','desktop','mobile','custom') NOT NULL DEFAULT 'custom',
  `is_kane_supported` tinyint(1) NOT NULL DEFAULT '0',
  `is_manual_supported` tinyint(1) NOT NULL DEFAULT '0',
  `is_default` tinyint(1) NOT NULL DEFAULT '0',
  `is_custom` tinyint(1) NOT NULL DEFAULT '0',
  `deleted_at` datetime DEFAULT NULL,
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `created_by` int NOT NULL,
  `updated_by` int NOT NULL,
  `is_complete` tinyint(1) DEFAULT '1',
  PRIMARY KEY (`id`),
  KEY `idx_configurations_organization_id` (`organization_id`),
  KEY `idx_configurations_name` (`name`),
  KEY `idx_configurations_is_kane_supported` (`is_kane_supported`),
  KEY `idx_configurations_is_manual_supported` (`is_manual_supported`),
  KEY `idx_configurations_is_default` (`is_default`),
  KEY `idx_configurations_platform` (`platform`)
) ENGINE=InnoDB AUTO_INCREMENT=104 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;



-- Altering the test_environments table to add a new column for configuration_id

ALTER TABLE `test_environments`
ADD COLUMN configuration_id BIGINT DEFAULT NULL,  -- removing for migration
ADD COLUMN udid VARCHAR(50) DEFAULT NULL,
ADD COLUMN platform_type VARCHAR(50) DEFAULT NULL,
ADD COLUMN metadata JSON DEFAULT NULL,
ADD CONSTRAINT fk_configuration_id
FOREIGN KEY (configuration_id)
REFERENCES configurations (id)
ON DELETE CASCADE;


--- Backing filling old data from child table test_environments to new table configurations
-- //  also link foreign key to test_environments table during backing filling


INSERT INTO configurations (id, organization_id, name, platform, is_kane_supported, is_manual_supported, is_default, is_custom, deleted_at, created_at, updated_at, created_by, updated_by, is_complete)
SELECT 
    id,
    organization_id,
    name,
    platform,
    is_kane_supported,
    is_manual_supported,
    is_default,
    is_custom,
    deleted_at,
    created_at,
    updated_at,
    created_by,
    updated_by,
    is_complete
FROM test_environments;


--  Update test_environments table to link foreign key to configurations table


SELECT DISTINCT organization_id
FROM test_environments;


