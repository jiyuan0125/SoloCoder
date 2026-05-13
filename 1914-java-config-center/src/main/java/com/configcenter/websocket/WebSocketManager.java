package com.configcenter.websocket;

import io.netty.channel.Channel;
import io.netty.channel.ChannelId;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;

public class WebSocketManager {

    private static final Logger logger = LoggerFactory.getLogger(WebSocketManager.class);

    private final Map<ChannelId, Channel> channels = new ConcurrentHashMap<>();
    private final Map<String, ChannelId> clientIdToChannelId = new ConcurrentHashMap<>();
    private final Map<ChannelId, String> channelIdToClientId = new ConcurrentHashMap<>();

    public void registerChannel(String clientId, Channel channel) {
        ChannelId channelId = channel.id();
        channels.put(channelId, channel);
        clientIdToChannelId.put(clientId, channelId);
        channelIdToClientId.put(channelId, clientId);
        logger.info("Client {} connected via WebSocket", clientId);
    }

    public void removeChannel(Channel channel) {
        ChannelId channelId = channel.id();
        channels.remove(channelId);
        String clientId = channelIdToClientId.remove(channelId);
        if (clientId != null) {
            clientIdToChannelId.remove(clientId);
            logger.info("Client {} disconnected from WebSocket", clientId);
        }
    }

    public Channel getChannel(String clientId) {
        ChannelId channelId = clientIdToChannelId.get(clientId);
        return channelId != null ? channels.get(channelId) : null;
    }

    public String getClientId(Channel channel) {
        return channelIdToClientId.get(channel.id());
    }

    public boolean isClientConnected(String clientId) {
        ChannelId channelId = clientIdToChannelId.get(clientId);
        if (channelId == null) return false;
        Channel channel = channels.get(channelId);
        return channel != null && channel.isActive();
    }

    public Map<ChannelId, Channel> getAllChannels() {
        return channels;
    }
}
