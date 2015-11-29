cancel($LIFDS_UPDATE_TIMER);

$LIFDS_ONLINE_UPDATE_INTERVAL = 10000; //Update list every 10 seconds

$LIFDSCP_nullObj = new SimObject();

package LIFDSCP
{
    function LIFDSCP_sqlNullCallback(%rs) 
    {
        dbi.remove(%rs);
        %rs.delete();
    }

    function LIFDSCP_sqlExecute(%sql)
    {
        echo("QUERY: " @ %sql);
        dbi.Execute(%sql, LIFDSCP_sqlNullCallback, $LIFDSCP_nullObj);
    }

    function LIFDSCP_onTimerElapsed()
    {
        echo("Updating" SPC ClientGroup.getCount() SPC "players...");
        %onlineChars = "";
        %haveChars = 0;
        LIFDSCP_sqlExecute("begin");

        for(%id = 0; %id < ClientGroup.getCount(); %id++)
        {
            %client = ClientGroup.getObject(%id);
            %pid = %client.getCharacterId();
            LIFDSCP_updatePlayer(%pid, %client);

            if (%id > 0)
            {
                %onlineChars = %onlineChars @ ", ";
            }
            
            %onlineChars = %onlineChars @ %pid;
            %haveChars = 1;
        }

        if (%haveChars > 0)
        {
            LIFDSCP_sqlExecute("delete from lifdscp_online_character where CharacterID not in (" @ %onlineChars @ ")");
        }
        else
        {
            LIFDSCP_sqlExecute("delete from lifdscp_online_character");
        }

        LIFDSCP_sqlExecute("commit");
        $LIFDS_UPDATE_TIMER = schedule($LIFDS_ONLINE_UPDATE_INTERVAL, 0, "LIFDSCP_onTimerElapsed");
    }

    function LIFDSCP_updatePlayer(%pid, %obj)
    {
        echo("Updating player " @ %pid);
        LIFDSCP_sqlExecute("insert ignore lifdscp_online_character (CharacterID) VALUES (" @ %pid @ ")");
    }
};

ActivatePackage(LIFDSCP);

$LIFDS_UPDATE_TIMER = schedule(100, 0, "LIFDSCP_onTimerElapsed");
