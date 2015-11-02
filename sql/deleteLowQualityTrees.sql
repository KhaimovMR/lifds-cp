drop procedure if exists p_deleteLowQualityTrees//
create procedure p_deleteLowQualityTrees()
begin
    declare v_i int unsigned default 442;
    declare v_quality int unsigned default 30;
    declare v_version int unsigned default 0;

    delete forest, forest_patch from forest, forest_patch where forest.GeoDataID = forest_patch.GeoDataID and forest.Quality < v_quality;

    while v_i < 451 do
        update forest_patch set Version = 2 where TerID = v_i;
        set v_version = (select Max(Version) as max from forest_patch where TerID = v_i);
        update terrain_blocks set ForestVersion = v_version where ID = v_i;
        set v_i = v_i + 1;
    end while;
end//
call p_deleteLowQualityTrees()
