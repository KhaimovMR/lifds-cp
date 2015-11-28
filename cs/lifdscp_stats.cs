cancel($UPDATE_TIMER);

$LIFDS_ONLINE_UPDATE_INTERVAL = 10000; //Update list every 10 seconds

$nullObj = new SimObject();

function sqlNullCallback(%rs) 
{
	dbi.remove(%rs);
	%rs.delete();
}

function sqlExecute(%sql)
{
    echo("QUERY: " @ %sql);
	dbi.Execute(%sql, sqlNullCallback, $nullObj);
}

function onTimerElapsed()
{
	echo("Updating" SPC ClientGroup.getCount() SPC "players...");
    %onlineChars = "";
    %haveChars = 0;
	sqlExecute("begin");

    for(%id = 0; %id < ClientGroup.getCount(); %id++)
    {
        %client = ClientGroup.getObject(%id);
        %pid = %client.getCharacterId();
        updatePlayer(%pid,%client);

        if (%id > 0)
        {
            %onlineChars = %onlineChars @ ", ";
        }
		
		%onlineChars = %onlineChars @ %pid;
        %haveChars = 1;
    }

    if (%haveChars > 0)
    {
        sqlExecute("delete from lifdscp_online_character where CharacterID not in (" @ %onlineChars @ ")");
    }
    else
    {
        sqlExecute("delete from lifdscp_online_character");
    }

    sqlExecute("commit");
	$UPDATE_TIMER = schedule($LIFDS_ONLINE_UPDATE_INTERVAL, 0, "onTimerElapsed");
}

function updatePlayer(%pid, %obj)
{
    echo("Updating player " @ %pid);
	sqlExecute("insert ignore lifdscp_online_character (CharacterID) VALUES (" @ %pid @ ")");
}

$UPDATE_TIMER = schedule(100, 0, "onTimerElapsed");
