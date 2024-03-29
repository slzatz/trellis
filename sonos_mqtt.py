#!bin/python
'''
Currently uses linode instance as mqtt server
Receives mqtt messages that initiate sonos actions
like shuffling an artist or playing a track or album
Uses the music services capabilities of SoCo
'''
import json
import paho.mqtt.client as mqtt
from soco.exceptions import SoCoSlaveException, SoCoUPnPException
from soco.music_services import MusicService
from soco.discovery import by_name
from config import mqtt_broker, music_service, speaker
from sonos_config import STATIONS, PLAYLISTS, META_FORMAT_PANDORA, META_FORMAT_RADIO
import random

topic = "trellis"
ms = MusicService(music_service)

master = by_name(speaker)
print("Master speaker is: {}".format(master.player_name))

for item in ms.get_metadata():
    if item.title == "My Music":
        items = ms.get_metadata(item.id)
        for my_item in ms.get_metadata(item.id):
            if my_item.title == "Playlists":
                #print("-> (" + type(my_item).__name__ + ")", my_item.title)
                playlists = ms.get_metadata(my_item.id)
                break
        break

pl_titles = [pl.title for pl in playlists]
pl_dict = dict(zip(pl_titles, playlists))

keyMap = {
		0:  "shuffle neil young",
		1:  "shuffle patty griffin",
		2:  "shuffle aimee mann",
		3:  "shuffle lucinda williams",
		4:  "shuffle jason isbell",
		5:  "shuffle radiohead",
		6:  "shuffle tom petty",
		7:  "shuffle amanda shires",
		8:  "shuffle jackson browne",
		9:  "shuffle ani difranco",
		10: "shuffle counting crows",
		11: "album after the goldrush neil young",
		12: "random_playlist",
		13: "station patty griffin",
		14: "next",
		15: "play_pause",
}

def on_connect(client, userdata, flags, rc):
    print("Connected with result code "+str(rc))

    # Subscribing in on_connect() means that if we lose the connection and
    # reconnect then subscriptions will be renewed.
    # client.subscribe("$SYS/#")
    client.subscribe(topic)

# The callback for when a PUBLISH message is received from the server.
def on_message(client, userdata, msg):
    #print(msg.topic+" "+str(msg.payload))
    print(f"topic: {msg.topic} payload: {msg.payload}")
    body = msg.payload
    #print("mqtt messge body =", body)

    try:
        key = json.loads(body).get('key', 15)
    except Exception as e:
        print("error reading the mqtt message body: ", e)
        return

    x = keyMap[key].split(" ", 1)
    if len(x) == 1:
        cmd = x[0]
    else:
        cmd, arg = x

    match cmd:

        case 'play_pause':
            try:
                state = master.get_current_transport_info()['current_transport_state']
            except Exception as e:
                print("Encountered error in state = master.get_current_transport_info(): ", e)
                state = 'ERROR'

            # check if sonos is playing music
            if state == 'PLAYING':
                master.pause()
            elif state!='ERROR':
                master.play()

        case 'play_queue':
            queue = master.get_queue()
            if len(queue) != 0:
                master.stop()
                master.play_from_queue(0)

        case 'quieter' | 'louder':
            for s in sp:
                s.volume = s.volume - 10 if cmd=='quieter' else s.volume + 10

            print("I tried to make the volume "+cmd)


        case 'next':
            try:
               master.next() 
            except SoCoUPnPException as e:
                print(f"master.{cmd}:", e)
            
            #sonos_actions2.playback('next')


        case 'station':
            station = STATIONS.get(arg.lower())
            if station:
                uri = station[1]
                if uri.startswith('x-sonosapi-radio'):
                    meta = META_FORMAT_PANDORA.format(title=station[0])
                elif uri.startswith('x-sonosapi-stream'):
                    meta = META_FORMAT_RADIO.format(title=station[0])

                master.play_uri(uri, meta, station[0]) # station[0] is the title of the station

            #sonos_actions2.play_station(arg)

        case 'shuffle':
            master.stop() # not necessary but let's you know a new cmd is underway
            master.clear_queue()
            tracks = []
            results = ms.search("tracks", arg)
            random.shuffle(results)
            # get something playing right away
            master.add_to_queue(results[0])
            master.play_from_queue(0)
            tracks.append(results[0].title)
            for track in list(results)[1:]:
                # Occasionally the artist is in some field search looks at but its not the artist for the song
                # may not be worth checking
                if track.title in tracks:
                    continue
                track_metadata = track.metadata.get('track_metadata', None)
                #print(f"{track_metadata.metadata.get('artist')=}: {arg=}")
                if not track_metadata:
                    continue
                if track_metadata.metadata.get('artist').lower() in arg: 
                    master.add_to_queue(track)
                    tracks.append(track.title)
            #master.play_from_queue(0)

        case 'album':
            master.stop()
            master.clear_queue()
            results = ms.search("albums", arg)
            master.add_to_queue(results[0])
            #master.play()    
            master.play_from_queue(0)

        case 'playlist':
            master.stop()
            master.clear_queue()
            pl = pl_dict[arg]
            master.add_to_queue(pl)
            #master.play()    
            master.play_from_queue(0)

        case 'random_playlist':
            master.stop()
            master.clear_queue()
            arg = random.choice(PLAYLISTS)
            pl = pl_dict[arg]
            master.add_to_queue(pl)
            #master.play()    
            master.play_from_queue(0)

        case _:
            print("I have no idea what you said")

#    action = task.get('action', '')
#
#    if action == 'play_pause':
#    
#        try:
#            state = master.get_current_transport_info()['current_transport_state']
#        except Exception as e:
#            print("Encountered error in state = master.get_current_transport_info(): ", e)
#            state = 'ERROR'
#
#        # check if sonos is playing music
#        if state == 'PLAYING':
#            master.pause()
#        elif state!='ERROR':
#            master.play()
#
#    elif action in ('quieter','louder'):
#        
#        for s in sp:
#            s.volume = s.volume - 10 if action=='quieter' else s.volume + 10
#
#        print("I tried to make the volume "+action)
#
#    elif action == "volume": #{"action":"volume", "level":70}
#
#        level = task.get("level", 500)
#        level = int(round(level/10, -1))
#
#        if level < 70:
#
#            for s in sp:
#                s.volume = level
#
#            print("I changed the volume to:", level)
#        else:
#            print("Volume was too high:", level)
#
#    #elif action == "play_wnyc":
#    #    sonos_actions.play_station('wnyc')
#
#    elif action == 'next':
#        sonos_actions.playback('next')
#
#    elif action.startswith("station"):
#        station = action[8:]
#        sonos_actions.play_station(station)
#
#    elif action.startswith("shuffle"):
#        master.clear_queue()
#        artist = action[8:]
#        results = ms.search("tracks", artist)
#        random.shuffle(results)
#        for track in results:
#            master.add_to_queue(track)
#        master.play()    
#
#    elif action.startswith("album"):
#        master.clear_queue()
#        album = action[6:]
#        results = ms.search("albums", album)
#        master.add_to_queue(results[0])
#        master.play()    
#
#    elif action.startswith("playlist"):
#        master.clear_queue()
#        pl_title = action[9:]
#        pl = pl_dict[pl_title]
#        master.add_to_queue(pl)
#        master.play()    
#    elif action.startswith("shuffle"):
#        artist = action[8:]
#        print(artist)
#       m = sonos_actions2.shuffle([artist])
#        print(m)
#
#    elif action == "play_queue":
#        queue = master.get_queue()
#        if len(queue) != 0:
#            master.stop()
#            master.play_from_queue(0)
#
#    else:
#        print("I have no idea what you said")

client = mqtt.Client()
client.on_connect = on_connect
client.on_message = on_message
client.connect(mqtt_broker, 1883, 60)
client.loop_forever()
