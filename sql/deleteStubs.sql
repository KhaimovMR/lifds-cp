DROP PROCEDURE IF EXISTS `p_deleteStubs`//
CREATE PROCEDURE `p_deleteStubs`()
BEGIN
   DECLARE done BOOLEAN DEFAULT FALSE;
   DECLARE cTerID, cVersion, cGeoID INT UNSIGNED;
   DECLARE cForest CURSOR FOR 
      SELECT filtered.`TerID`, MAX(filtered.`Version`) AS `Version`, filtered.`GeoDataID`
      FROM `forest_patch` AS filtered
      INNER JOIN(
         SELECT `TerID`, MAX(`Version`) AS `Version`, `GeoDataID` 
         FROM `forest_patch`
         WHERE `Action` !=4
         GROUP BY `GeoDataID`
      ) AS actual
      ON actual.`GeoDataID` = filtered.`GeoDataID` AND actual.`Version` < filtered.`Version`
      WHERE filtered.`Action` =4 
      GROUP BY filtered.`GeoDataID`;
   DECLARE CONTINUE HANDLER FOR NOT FOUND SET done = TRUE;
   
   OPEN cForest;
   
   trunk_loop: LOOP
      FETCH cForest INTO cTerID, cVersion, cGeoID;
      IF done THEN 
         LEAVE trunk_loop;
      END IF;
      set @dummy_var = (SELECT f_addForestPatch(cTerID,3,cGeoID,NULL,NULL,NULL,NULL));
      CALL p_deleteForestItem(cGeoID);
   END LOOP;
   
   CLOSE cForest;
END//
CALL p_deleteStubs()
