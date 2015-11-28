cancel($LIFDS_UPDATE_TIMER);

$LIFDS_ONLINE_UPDATE_INTERVAL = 10000; //Update list every 10 seconds

$LIFDS_nullObj = new SimObject();

package LiFDSCP
{

function LIFDS_sqlNullCallback(%rs) 
{
    dbi.remove(%rs);
    %rs.delete();
}

function LIFDS_sqlExecute(%sql)
{
    echo("QUERY: " @ %sql);
    dbi.Execute(%sql, LIFDS_sqlNullCallback, $LIFDS_nullObj);
}

function LIFDS_onTimerElapsed()
{
    echo("Updating" SPC ClientGroup.getCount() SPC "players...");
    %onlineChars = "";
    %haveChars = 0;
    sqlExecute("begin");

    for(%id = 0; %id < ClientGroup.getCount(); %id++)
    {
        %client = ClientGroup.getObject(%id);
        %pid = %client.getCharacterId();
        LIFDS_updatePlayer(%pid,%client);

        if (%id > 0)
        {
            %onlineChars = %onlineChars @ ", ";
        }
		
        %onlineChars = %onlineChars @ %pid;
        %haveChars = 1;
    }

    if (%haveChars > 0)
    {
        LIFDS_sqlExecute("delete from lifdscp_online_character where CharacterID not in (" @ %onlineChars @ ")");
    }
    else
    {
        LIFDS_sqlExecute("delete from lifdscp_online_character");
    }

    LIFDS_sqlExecute("commit");
    $LIFDS_UPDATE_TIMER = schedule($LIFDS_ONLINE_UPDATE_INTERVAL, 0, "LIFDS_onTimerElapsed");
}

function LIFDS_updatePlayer(%pid, %obj)
{
    echo("Updating player " @ %pid);
    LIFDS_sqlExecute("insert ignore lifdscp_online_character (CharacterID) VALUES (" @ %pid @ ")");
}

};

ActivatePackage(LiFDSCP);

$LIFDS_UPDATE_TIMER = schedule(100, 0, "LIFDS_onTimerElapsed");
