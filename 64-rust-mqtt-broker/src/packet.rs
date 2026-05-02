use std::io::{Read, Write, Error, ErrorKind};

#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum PacketType {
    Connect = 1,
    Connack = 2,
    Publish = 3,
    Puback = 4,
    Pubrec = 5,
    Pubrel = 6,
    Pubcomp = 7,
    Subscribe = 8,
    Suback = 9,
    Unsubscribe = 10,
    Unsuback = 11,
    Pingreq = 12,
    Pingresp = 13,
    Disconnect = 14,
}

impl PacketType {
    pub fn from_u8(value: u8) -> Option<Self> {
        match value {
            1 => Some(PacketType::Connect),
            2 => Some(PacketType::Connack),
            3 => Some(PacketType::Publish),
            4 => Some(PacketType::Puback),
            5 => Some(PacketType::Pubrec),
            6 => Some(PacketType::Pubrel),
            7 => Some(PacketType::Pubcomp),
            8 => Some(PacketType::Subscribe),
            9 => Some(PacketType::Suback),
            10 => Some(PacketType::Unsubscribe),
            11 => Some(PacketType::Unsuback),
            12 => Some(PacketType::Pingreq),
            13 => Some(PacketType::Pingresp),
            14 => Some(PacketType::Disconnect),
            _ => None,
        }
    }
}

#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum QoS {
    AtMostOnce = 0,
    AtLeastOnce = 1,
    ExactlyOnce = 2,
}

impl QoS {
    pub fn from_u8(value: u8) -> Option<Self> {
        match value {
            0 => Some(QoS::AtMostOnce),
            1 => Some(QoS::AtLeastOnce),
            2 => Some(QoS::ExactlyOnce),
            _ => None,
        }
    }
}

#[derive(Debug, Clone, PartialEq, Eq)]
pub struct FixedHeader {
    pub packet_type: PacketType,
    pub dup: bool,
    pub qos: QoS,
    pub retain: bool,
    pub remaining_length: usize,
}

#[derive(Debug, Clone, PartialEq, Eq)]
pub struct ConnectFlags {
    pub username_flag: bool,
    pub password_flag: bool,
    pub will_retain: bool,
    pub will_qos: QoS,
    pub will_flag: bool,
    pub clean_session: bool,
}

#[derive(Debug, Clone, PartialEq, Eq)]
pub struct ConnectPacket {
    pub protocol_name: String,
    pub protocol_level: u8,
    pub connect_flags: ConnectFlags,
    pub keep_alive: u16,
    pub client_id: String,
    pub will_topic: Option<String>,
    pub will_message: Option<Vec<u8>>,
    pub username: Option<String>,
    pub password: Option<Vec<u8>>,
}

#[derive(Debug, Clone, PartialEq, Eq)]
pub struct ConnackPacket {
    pub session_present: bool,
    pub return_code: u8,
}

#[derive(Debug, Clone, PartialEq, Eq)]
pub struct PublishPacket {
    pub topic_name: String,
    pub packet_identifier: Option<u16>,
    pub payload: Vec<u8>,
}

#[derive(Debug, Clone, PartialEq, Eq)]
pub struct PubackPacket {
    pub packet_identifier: u16,
}

#[derive(Debug, Clone, PartialEq, Eq)]
pub struct SubscribeTopic {
    pub topic_filter: String,
    pub qos: QoS,
}

#[derive(Debug, Clone, PartialEq, Eq)]
pub struct SubscribePacket {
    pub packet_identifier: u16,
    pub topics: Vec<SubscribeTopic>,
}

#[derive(Debug, Clone, PartialEq, Eq)]
pub struct SubackPacket {
    pub packet_identifier: u16,
    pub return_codes: Vec<u8>,
}

#[derive(Debug, Clone, PartialEq, Eq)]
pub struct UnsubscribePacket {
    pub packet_identifier: u16,
    pub topics: Vec<String>,
}

#[derive(Debug, Clone, PartialEq, Eq)]
pub struct UnsubackPacket {
    pub packet_identifier: u16,
}

#[derive(Debug, Clone, PartialEq, Eq)]
pub enum Packet {
    Connect(ConnectPacket),
    Connack(ConnackPacket),
    Publish(PublishPacket),
    Puback(PubackPacket),
    Subscribe(SubscribePacket),
    Suback(SubackPacket),
    Unsubscribe(UnsubscribePacket),
    Unsuback(UnsubackPacket),
    Pingreq,
    Pingresp,
    Disconnect,
}

pub fn read_u16(reader: &mut dyn Read) -> Result<u16, Error> {
    let mut buf = [0u8; 2];
    reader.read_exact(&mut buf)?;
    Ok(((buf[0] as u16) << 8) | (buf[1] as u16))
}

pub fn write_u16(writer: &mut dyn Write, value: u16) -> Result<(), Error> {
    let buf = [((value >> 8) & 0xFF) as u8, (value & 0xFF) as u8];
    writer.write_all(&buf)
}

pub fn read_utf8_string(reader: &mut dyn Read) -> Result<String, Error> {
    let len = read_u16(reader)?;
    let mut buf = vec![0u8; len as usize];
    reader.read_exact(&mut buf)?;
    String::from_utf8(buf).map_err(|_| Error::new(ErrorKind::InvalidData, "Invalid UTF-8"))
}

pub fn write_utf8_string(writer: &mut dyn Write, s: &str) -> Result<(), Error> {
    write_u16(writer, s.len() as u16)?;
    writer.write_all(s.as_bytes())
}

pub fn read_remaining_length(reader: &mut dyn Read) -> Result<usize, Error> {
    let mut multiplier = 1;
    let mut value = 0usize;
    loop {
        let mut buf = [0u8; 1];
        reader.read_exact(&mut buf)?;
        let byte = buf[0];
        value += ((byte & 0x7F) as usize) * multiplier;
        if multiplier > 128 * 128 * 128 {
            return Err(Error::new(ErrorKind::InvalidData, "Malformed remaining length"));
        }
        multiplier *= 128;
        if (byte & 0x80) == 0 {
            break;
        }
    }
    Ok(value)
}

pub fn write_remaining_length(writer: &mut dyn Write, mut length: usize) -> Result<(), Error> {
    loop {
        let mut byte = (length % 128) as u8;
        length = length / 128;
        if length > 0 {
            byte |= 0x80;
        }
        writer.write_all(&[byte])?;
        if length == 0 {
            break;
        }
    }
    Ok(())
}

pub fn remaining_length_bytes(mut length: usize) -> usize {
    let mut count = 0;
    loop {
        length /= 128;
        count += 1;
        if length == 0 {
            break;
        }
    }
    count
}

pub fn parse_fixed_header(first_byte: u8, remaining_length: usize) -> Result<FixedHeader, Error> {
    let packet_type_value = (first_byte >> 4) & 0x0F;
    let packet_type = PacketType::from_u8(packet_type_value)
        .ok_or_else(|| Error::new(ErrorKind::InvalidData, "Invalid packet type"))?;
    
    let dup = ((first_byte >> 3) & 0x01) != 0;
    let qos_value = (first_byte >> 1) & 0x03;
    let qos = QoS::from_u8(qos_value)
        .ok_or_else(|| Error::new(ErrorKind::InvalidData, "Invalid QoS"))?;
    let retain = (first_byte & 0x01) != 0;

    Ok(FixedHeader {
        packet_type,
        dup,
        qos,
        retain,
        remaining_length,
    })
}

pub fn build_fixed_header_byte(fixed_header: &FixedHeader) -> u8 {
    let mut byte = (fixed_header.packet_type as u8) << 4;
    if fixed_header.dup {
        byte |= 0x08;
    }
    byte |= (fixed_header.qos as u8) << 1;
    if fixed_header.retain {
        byte |= 0x01;
    }
    byte
}

pub fn parse_connect_flags(flags_byte: u8) -> Result<ConnectFlags, Error> {
    let username_flag = (flags_byte & 0x80) != 0;
    let password_flag = (flags_byte & 0x40) != 0;
    let will_retain = (flags_byte & 0x20) != 0;
    let will_qos_value = (flags_byte >> 3) & 0x03;
    let will_qos = QoS::from_u8(will_qos_value)
        .ok_or_else(|| Error::new(ErrorKind::InvalidData, "Invalid will QoS"))?;
    let will_flag = (flags_byte & 0x04) != 0;
    let clean_session = (flags_byte & 0x02) != 0;

    Ok(ConnectFlags {
        username_flag,
        password_flag,
        will_retain,
        will_qos,
        will_flag,
        clean_session,
    })
}

pub fn build_connect_flags(flags: &ConnectFlags) -> u8 {
    let mut byte = 0u8;
    if flags.username_flag {
        byte |= 0x80;
    }
    if flags.password_flag {
        byte |= 0x40;
    }
    if flags.will_retain {
        byte |= 0x20;
    }
    byte |= (flags.will_qos as u8) << 3;
    if flags.will_flag {
        byte |= 0x04;
    }
    if flags.clean_session {
        byte |= 0x02;
    }
    byte
}

pub fn parse_connect_packet(reader: &mut dyn Read, remaining_length: usize) -> Result<ConnectPacket, Error> {
    let mut buf = vec![0u8; remaining_length];
    reader.read_exact(&mut buf)?;
    let mut cursor = std::io::Cursor::new(&buf);
    
    let protocol_name = read_utf8_string(&mut cursor)?;
    let mut level_buf = [0u8; 1];
    cursor.read_exact(&mut level_buf)?;
    let protocol_level = level_buf[0];
    
    let mut flags_buf = [0u8; 1];
    cursor.read_exact(&mut flags_buf)?;
    let connect_flags = parse_connect_flags(flags_buf[0])?;
    
    let keep_alive = read_u16(&mut cursor)?;
    let client_id = read_utf8_string(&mut cursor)?;
    
    let will_topic = if connect_flags.will_flag {
        Some(read_utf8_string(&mut cursor)?)
    } else {
        None
    };
    
    let will_message = if connect_flags.will_flag {
        let len = read_u16(&mut cursor)?;
        let mut msg_buf = vec![0u8; len as usize];
        cursor.read_exact(&mut msg_buf)?;
        Some(msg_buf)
    } else {
        None
    };
    
    let username = if connect_flags.username_flag {
        Some(read_utf8_string(&mut cursor)?)
    } else {
        None
    };
    
    let password = if connect_flags.password_flag {
        let len = read_u16(&mut cursor)?;
        let mut pwd_buf = vec![0u8; len as usize];
        cursor.read_exact(&mut pwd_buf)?;
        Some(pwd_buf)
    } else {
        None
    };

    Ok(ConnectPacket {
        protocol_name,
        protocol_level,
        connect_flags,
        keep_alive,
        client_id,
        will_topic,
        will_message,
        username,
        password,
    })
}

pub fn encode_connack_packet(packet: &ConnackPacket) -> Result<Vec<u8>, Error> {
    let remaining_length = 2;
    let mut buf = Vec::with_capacity(1 + remaining_length_bytes(remaining_length) + remaining_length);
    
    let fixed_header = FixedHeader {
        packet_type: PacketType::Connack,
        dup: false,
        qos: QoS::AtMostOnce,
        retain: false,
        remaining_length,
    };
    buf.push(build_fixed_header_byte(&fixed_header));
    write_remaining_length(&mut buf, remaining_length)?;
    
    let connect_ack_flags = if packet.session_present { 0x01 } else { 0x00 };
    buf.push(connect_ack_flags);
    buf.push(packet.return_code);
    
    Ok(buf)
}

pub fn parse_publish_packet(fixed_header: &FixedHeader, reader: &mut dyn Read) -> Result<PublishPacket, Error> {
    let mut buf = vec![0u8; fixed_header.remaining_length];
    reader.read_exact(&mut buf)?;
    let mut cursor = std::io::Cursor::new(&buf);
    
    let topic_name = read_utf8_string(&mut cursor)?;
    
    let packet_identifier = if fixed_header.qos != QoS::AtMostOnce {
        Some(read_u16(&mut cursor)?)
    } else {
        None
    };
    
    let payload_start = cursor.position() as usize;
    let payload = buf[payload_start..].to_vec();

    Ok(PublishPacket {
        topic_name,
        packet_identifier,
        payload,
    })
}

pub fn encode_publish_packet(
    topic: &str,
    payload: &[u8],
    qos: QoS,
    retain: bool,
    dup: bool,
    packet_id: Option<u16>,
) -> Result<Vec<u8>, Error> {
    let mut variable_header_len = 2 + topic.len();
    if qos != QoS::AtMostOnce && packet_id.is_some() {
        variable_header_len += 2;
    }
    let remaining_length = variable_header_len + payload.len();
    
    let mut buf = Vec::with_capacity(1 + remaining_length_bytes(remaining_length) + remaining_length);
    
    let fixed_header = FixedHeader {
        packet_type: PacketType::Publish,
        dup,
        qos,
        retain,
        remaining_length,
    };
    buf.push(build_fixed_header_byte(&fixed_header));
    write_remaining_length(&mut buf, remaining_length)?;
    
    write_utf8_string(&mut buf, topic)?;
    if qos != QoS::AtMostOnce {
        if let Some(id) = packet_id {
            write_u16(&mut buf, id)?;
        }
    }
    buf.write_all(payload)?;
    
    Ok(buf)
}

pub fn parse_puback_packet(reader: &mut dyn Read) -> Result<PubackPacket, Error> {
    let mut buf = [0u8; 2];
    reader.read_exact(&mut buf)?;
    let packet_identifier = ((buf[0] as u16) << 8) | (buf[1] as u16);
    Ok(PubackPacket { packet_identifier })
}

pub fn encode_puback_packet(packet_id: u16) -> Result<Vec<u8>, Error> {
    let remaining_length = 2;
    let mut buf = Vec::with_capacity(1 + remaining_length_bytes(remaining_length) + remaining_length);
    
    let fixed_header = FixedHeader {
        packet_type: PacketType::Puback,
        dup: false,
        qos: QoS::AtMostOnce,
        retain: false,
        remaining_length,
    };
    buf.push(build_fixed_header_byte(&fixed_header));
    write_remaining_length(&mut buf, remaining_length)?;
    write_u16(&mut buf, packet_id)?;
    
    Ok(buf)
}

pub fn parse_subscribe_packet(reader: &mut dyn Read, remaining_length: usize) -> Result<SubscribePacket, Error> {
    let mut buf = vec![0u8; remaining_length];
    reader.read_exact(&mut buf)?;
    let mut cursor = std::io::Cursor::new(&buf);
    
    let packet_identifier = read_u16(&mut cursor)?;
    let mut topics = Vec::new();
    
    while cursor.position() < remaining_length as u64 {
        let topic_filter = read_utf8_string(&mut cursor)?;
        let mut qos_buf = [0u8; 1];
        cursor.read_exact(&mut qos_buf)?;
        let qos = QoS::from_u8(qos_buf[0] & 0x03)
            .ok_or_else(|| Error::new(ErrorKind::InvalidData, "Invalid QoS in SUBSCRIBE"))?;
        topics.push(SubscribeTopic { topic_filter, qos });
    }

    Ok(SubscribePacket {
        packet_identifier,
        topics,
    })
}

pub fn encode_suback_packet(packet_id: u16, return_codes: &[u8]) -> Result<Vec<u8>, Error> {
    let remaining_length = 2 + return_codes.len();
    let mut buf = Vec::with_capacity(1 + remaining_length_bytes(remaining_length) + remaining_length);
    
    let fixed_header = FixedHeader {
        packet_type: PacketType::Suback,
        dup: false,
        qos: QoS::AtMostOnce,
        retain: false,
        remaining_length,
    };
    buf.push(build_fixed_header_byte(&fixed_header));
    write_remaining_length(&mut buf, remaining_length)?;
    write_u16(&mut buf, packet_id)?;
    buf.write_all(return_codes)?;
    
    Ok(buf)
}

pub fn parse_unsubscribe_packet(reader: &mut dyn Read, remaining_length: usize) -> Result<UnsubscribePacket, Error> {
    let mut buf = vec![0u8; remaining_length];
    reader.read_exact(&mut buf)?;
    let mut cursor = std::io::Cursor::new(&buf);
    
    let packet_identifier = read_u16(&mut cursor)?;
    let mut topics = Vec::new();
    
    while cursor.position() < remaining_length as u64 {
        let topic_filter = read_utf8_string(&mut cursor)?;
        topics.push(topic_filter);
    }

    Ok(UnsubscribePacket {
        packet_identifier,
        topics,
    })
}

pub fn encode_unsuback_packet(packet_id: u16) -> Result<Vec<u8>, Error> {
    let remaining_length = 2;
    let mut buf = Vec::with_capacity(1 + remaining_length_bytes(remaining_length) + remaining_length);
    
    let fixed_header = FixedHeader {
        packet_type: PacketType::Unsuback,
        dup: false,
        qos: QoS::AtMostOnce,
        retain: false,
        remaining_length,
    };
    buf.push(build_fixed_header_byte(&fixed_header));
    write_remaining_length(&mut buf, remaining_length)?;
    write_u16(&mut buf, packet_id)?;
    
    Ok(buf)
}

pub fn encode_pingresp_packet() -> Result<Vec<u8>, Error> {
    let remaining_length = 0;
    let mut buf = Vec::with_capacity(1 + remaining_length_bytes(remaining_length));
    
    let fixed_header = FixedHeader {
        packet_type: PacketType::Pingresp,
        dup: false,
        qos: QoS::AtMostOnce,
        retain: false,
        remaining_length,
    };
    buf.push(build_fixed_header_byte(&fixed_header));
    write_remaining_length(&mut buf, remaining_length)?;
    
    Ok(buf)
}

pub fn read_packet(reader: &mut dyn Read) -> Result<(FixedHeader, Packet), Error> {
    let mut first_byte_buf = [0u8; 1];
    match reader.read_exact(&mut first_byte_buf) {
        Ok(_) => {}
        Err(e) if e.kind() == ErrorKind::UnexpectedEof => {
            return Err(Error::new(ErrorKind::ConnectionAborted, "Connection closed"));
        }
        Err(e) => return Err(e),
    }
    let first_byte = first_byte_buf[0];
    
    let remaining_length = read_remaining_length(reader)?;
    let fixed_header = parse_fixed_header(first_byte, remaining_length)?;

    let packet = match fixed_header.packet_type {
        PacketType::Connect => {
            let pkt = parse_connect_packet(reader, remaining_length)?;
            Packet::Connect(pkt)
        }
        PacketType::Publish => {
            let pkt = parse_publish_packet(&fixed_header, reader)?;
            Packet::Publish(pkt)
        }
        PacketType::Puback => {
            let pkt = parse_puback_packet(reader)?;
            Packet::Puback(pkt)
        }
        PacketType::Subscribe => {
            let pkt = parse_subscribe_packet(reader, remaining_length)?;
            Packet::Subscribe(pkt)
        }
        PacketType::Unsubscribe => {
            let pkt = parse_unsubscribe_packet(reader, remaining_length)?;
            Packet::Unsubscribe(pkt)
        }
        PacketType::Pingreq => {
            Packet::Pingreq
        }
        PacketType::Disconnect => {
            Packet::Disconnect
        }
        _ => {
            return Err(Error::new(ErrorKind::InvalidData, "Unsupported packet type"));
        }
    };

    Ok((fixed_header, packet))
}
