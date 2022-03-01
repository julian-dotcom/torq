/* eslint-disable */
import Long from "long";
import { grpc } from "@improbable-eng/grpc-web";
import _m0 from "protobufjs/minimal";
import { BrowserHeaders } from "browser-headers";

export const protobufPackage = "torqrpc";

export enum GroupType {
  CHANNEL = 0,
  PEER = 1,
  TAG = 2,
  UNRECOGNIZED = -1,
}

export function groupTypeFromJSON(object: any): GroupType {
  switch (object) {
    case 0:
    case "CHANNEL":
      return GroupType.CHANNEL;
    case 1:
    case "PEER":
      return GroupType.PEER;
    case 2:
    case "TAG":
      return GroupType.TAG;
    case -1:
    case "UNRECOGNIZED":
    default:
      return GroupType.UNRECOGNIZED;
  }
}

export function groupTypeToJSON(object: GroupType): string {
  switch (object) {
    case GroupType.CHANNEL:
      return "CHANNEL";
    case GroupType.PEER:
      return "PEER";
    case GroupType.TAG:
      return "TAG";
    default:
      return "UNKNOWN";
  }
}

/**
 * ------------------------------------------------------------------
 * Aggregated forwards
 * ------------------------------------------------------------------
 */
export interface AggregatedForwardsRequest {
  /** From what date/time (unix timestamp) */
  fromTs: Long;
  /** To what date/time (unix timestamp) */
  toTs: Long;
  channelIds: ChannelIDs | undefined;
  peerIds: PeerIDs | undefined;
  tagIds: TagIDs | undefined;
}

export interface AggregatedForwardsResponse {
  /** From what date/time (unix timestamp) */
  fromTs: Long;
  /** To what date/time (unix timestamp) */
  toTs: Long;
  /**
   * This represents the type of aggregation
   * 0 = CHANNEL, means forwards are aggregated per channel
   * 1 = PEER, means forwards are aggregated per peer (remote public key)
   * 1 = TAG, means forwards are aggregated per tag, tags are a user defined group of channels.
   */
  groupType: GroupType;
  /** a list of aggregated forwarding stats */
  aggregatedForwards: AggregatedForwards[];
}

/** AggregatedForwards represents a channels, peers or tags aggregated forwarding stats. */
export interface AggregatedForwards {
  /** A list of the channel included in the aggregated result. */
  channels: ChanInfo[];
  /**
   * This represents the type of aggregation
   * 0 = CHANNEL, means forwards are aggregated per channel
   * 1 = PEER, means forwards are aggregated per peer (remote public key)
   * 1 = TAG, means forwards are aggregated per tag, tags are a user defined group of channels.
   */
  groupType: GroupType;
  /**
   * The group_id depends on the granularity requested.
   *   * Channel ID for individual channels
   *   * Remote peers public key when it's grouped per peer
   *   * Tag id when it's grouped by tag.
   */
  groupId: string;
  /**
   * The group_name depends on the granularity requested.
   *   * Individual channels use the format: <Alias> (<Channel ID>)
   *   * Peer grouped use the format: <Alias>
   *   * Tag groups use the format: <Tag Name>
   */
  groupName: string;
  /** The  utbound amount in sats (Satoshis) */
  amountOut: Long;
  /** The inbound amount in sats (Satoshis) */
  amountIn: Long;
  /** The total amount forwarded both inbound and outbound */
  amountTotal: Long;
  /** The outbound revenue in sats. This is what the channel has directly produced. */
  revenueOut: Long;
  /**
   * The inbound revenue in sats. This is what the channel has indirectly produced.
   * This revenue are not really earned by this channel/peer/group, but represents
   * the channel/peer/group contribution to revenue earned by other channels.
   */
  revenueIn: Long;
  /** Total revenue */
  revenueTotal: Long;
  /** Number of outbound forwards. */
  countOut: Long;
  /** Number of inbound forwards. */
  countIn: Long;
  /** Number of inbound forwards. */
  countTotal: Long;
  /** The total capacity of this channel */
  capacity: Long;
  /**
   * Turnover = capacity/amount_total. It is a measure of how many times
   * the capital is used in the given time interval.
   */
  turnover: number;
}

export interface ChannelIDs {
  /** A list of channel IDs */
  chanIds: Long[];
}

export interface PeerIDs {
  /** A list of public kees belonging to peers you have channels with */
  pubKeys: string[];
}

export interface TagIDs {
  /** A list of tag ids */
  tagIds: string[];
}

export interface ChanInfo {
  /** The channel ID */
  chanId: Long;
  /** Alias of remote peer */
  alias: string;
  /** The remote public key */
  pubKey: string;
  /** True if the channel is closed */
  closed: boolean;
  /** A list of tags associated with this channel */
  tags: string[];
}

const baseAggregatedForwardsRequest: object = {
  fromTs: Long.ZERO,
  toTs: Long.ZERO,
};

export const AggregatedForwardsRequest = {
  encode(
    message: AggregatedForwardsRequest,
    writer: _m0.Writer = _m0.Writer.create()
  ): _m0.Writer {
    if (!message.fromTs.isZero()) {
      writer.uint32(8).int64(message.fromTs);
    }
    if (!message.toTs.isZero()) {
      writer.uint32(16).int64(message.toTs);
    }
    if (message.channelIds !== undefined) {
      ChannelIDs.encode(message.channelIds, writer.uint32(26).fork()).ldelim();
    }
    if (message.peerIds !== undefined) {
      PeerIDs.encode(message.peerIds, writer.uint32(34).fork()).ldelim();
    }
    if (message.tagIds !== undefined) {
      TagIDs.encode(message.tagIds, writer.uint32(42).fork()).ldelim();
    }
    return writer;
  },

  decode(
    input: _m0.Reader | Uint8Array,
    length?: number
  ): AggregatedForwardsRequest {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseAggregatedForwardsRequest,
    } as AggregatedForwardsRequest;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.fromTs = reader.int64() as Long;
          break;
        case 2:
          message.toTs = reader.int64() as Long;
          break;
        case 3:
          message.channelIds = ChannelIDs.decode(reader, reader.uint32());
          break;
        case 4:
          message.peerIds = PeerIDs.decode(reader, reader.uint32());
          break;
        case 5:
          message.tagIds = TagIDs.decode(reader, reader.uint32());
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): AggregatedForwardsRequest {
    const message = {
      ...baseAggregatedForwardsRequest,
    } as AggregatedForwardsRequest;
    message.fromTs =
      object.fromTs !== undefined && object.fromTs !== null
        ? Long.fromString(object.fromTs)
        : Long.ZERO;
    message.toTs =
      object.toTs !== undefined && object.toTs !== null
        ? Long.fromString(object.toTs)
        : Long.ZERO;
    message.channelIds =
      object.channelIds !== undefined && object.channelIds !== null
        ? ChannelIDs.fromJSON(object.channelIds)
        : undefined;
    message.peerIds =
      object.peerIds !== undefined && object.peerIds !== null
        ? PeerIDs.fromJSON(object.peerIds)
        : undefined;
    message.tagIds =
      object.tagIds !== undefined && object.tagIds !== null
        ? TagIDs.fromJSON(object.tagIds)
        : undefined;
    return message;
  },

  toJSON(message: AggregatedForwardsRequest): unknown {
    const obj: any = {};
    message.fromTs !== undefined &&
      (obj.fromTs = (message.fromTs || Long.ZERO).toString());
    message.toTs !== undefined &&
      (obj.toTs = (message.toTs || Long.ZERO).toString());
    message.channelIds !== undefined &&
      (obj.channelIds = message.channelIds
        ? ChannelIDs.toJSON(message.channelIds)
        : undefined);
    message.peerIds !== undefined &&
      (obj.peerIds = message.peerIds
        ? PeerIDs.toJSON(message.peerIds)
        : undefined);
    message.tagIds !== undefined &&
      (obj.tagIds = message.tagIds ? TagIDs.toJSON(message.tagIds) : undefined);
    return obj;
  },

  fromPartial<I extends Exact<DeepPartial<AggregatedForwardsRequest>, I>>(
    object: I
  ): AggregatedForwardsRequest {
    const message = {
      ...baseAggregatedForwardsRequest,
    } as AggregatedForwardsRequest;
    message.fromTs =
      object.fromTs !== undefined && object.fromTs !== null
        ? Long.fromValue(object.fromTs)
        : Long.ZERO;
    message.toTs =
      object.toTs !== undefined && object.toTs !== null
        ? Long.fromValue(object.toTs)
        : Long.ZERO;
    message.channelIds =
      object.channelIds !== undefined && object.channelIds !== null
        ? ChannelIDs.fromPartial(object.channelIds)
        : undefined;
    message.peerIds =
      object.peerIds !== undefined && object.peerIds !== null
        ? PeerIDs.fromPartial(object.peerIds)
        : undefined;
    message.tagIds =
      object.tagIds !== undefined && object.tagIds !== null
        ? TagIDs.fromPartial(object.tagIds)
        : undefined;
    return message;
  },
};

const baseAggregatedForwardsResponse: object = {
  fromTs: Long.ZERO,
  toTs: Long.ZERO,
  groupType: 0,
};

export const AggregatedForwardsResponse = {
  encode(
    message: AggregatedForwardsResponse,
    writer: _m0.Writer = _m0.Writer.create()
  ): _m0.Writer {
    if (!message.fromTs.isZero()) {
      writer.uint32(8).int64(message.fromTs);
    }
    if (!message.toTs.isZero()) {
      writer.uint32(16).int64(message.toTs);
    }
    if (message.groupType !== 0) {
      writer.uint32(24).int32(message.groupType);
    }
    for (const v of message.aggregatedForwards) {
      AggregatedForwards.encode(v!, writer.uint32(34).fork()).ldelim();
    }
    return writer;
  },

  decode(
    input: _m0.Reader | Uint8Array,
    length?: number
  ): AggregatedForwardsResponse {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseAggregatedForwardsResponse,
    } as AggregatedForwardsResponse;
    message.aggregatedForwards = [];
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.fromTs = reader.int64() as Long;
          break;
        case 2:
          message.toTs = reader.int64() as Long;
          break;
        case 3:
          message.groupType = reader.int32() as any;
          break;
        case 4:
          message.aggregatedForwards.push(
            AggregatedForwards.decode(reader, reader.uint32())
          );
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): AggregatedForwardsResponse {
    const message = {
      ...baseAggregatedForwardsResponse,
    } as AggregatedForwardsResponse;
    message.fromTs =
      object.fromTs !== undefined && object.fromTs !== null
        ? Long.fromString(object.fromTs)
        : Long.ZERO;
    message.toTs =
      object.toTs !== undefined && object.toTs !== null
        ? Long.fromString(object.toTs)
        : Long.ZERO;
    message.groupType =
      object.groupType !== undefined && object.groupType !== null
        ? groupTypeFromJSON(object.groupType)
        : 0;
    message.aggregatedForwards = (object.aggregatedForwards ?? []).map(
      (e: any) => AggregatedForwards.fromJSON(e)
    );
    return message;
  },

  toJSON(message: AggregatedForwardsResponse): unknown {
    const obj: any = {};
    message.fromTs !== undefined &&
      (obj.fromTs = (message.fromTs || Long.ZERO).toString());
    message.toTs !== undefined &&
      (obj.toTs = (message.toTs || Long.ZERO).toString());
    message.groupType !== undefined &&
      (obj.groupType = groupTypeToJSON(message.groupType));
    if (message.aggregatedForwards) {
      obj.aggregatedForwards = message.aggregatedForwards.map((e) =>
        e ? AggregatedForwards.toJSON(e) : undefined
      );
    } else {
      obj.aggregatedForwards = [];
    }
    return obj;
  },

  fromPartial<I extends Exact<DeepPartial<AggregatedForwardsResponse>, I>>(
    object: I
  ): AggregatedForwardsResponse {
    const message = {
      ...baseAggregatedForwardsResponse,
    } as AggregatedForwardsResponse;
    message.fromTs =
      object.fromTs !== undefined && object.fromTs !== null
        ? Long.fromValue(object.fromTs)
        : Long.ZERO;
    message.toTs =
      object.toTs !== undefined && object.toTs !== null
        ? Long.fromValue(object.toTs)
        : Long.ZERO;
    message.groupType = object.groupType ?? 0;
    message.aggregatedForwards =
      object.aggregatedForwards?.map((e) =>
        AggregatedForwards.fromPartial(e)
      ) || [];
    return message;
  },
};

const baseAggregatedForwards: object = {
  groupType: 0,
  groupId: "",
  groupName: "",
  amountOut: Long.UZERO,
  amountIn: Long.UZERO,
  amountTotal: Long.UZERO,
  revenueOut: Long.UZERO,
  revenueIn: Long.UZERO,
  revenueTotal: Long.UZERO,
  countOut: Long.UZERO,
  countIn: Long.UZERO,
  countTotal: Long.UZERO,
  capacity: Long.UZERO,
  turnover: 0,
};

export const AggregatedForwards = {
  encode(
    message: AggregatedForwards,
    writer: _m0.Writer = _m0.Writer.create()
  ): _m0.Writer {
    for (const v of message.channels) {
      ChanInfo.encode(v!, writer.uint32(10).fork()).ldelim();
    }
    if (message.groupType !== 0) {
      writer.uint32(16).int32(message.groupType);
    }
    if (message.groupId !== "") {
      writer.uint32(34).string(message.groupId);
    }
    if (message.groupName !== "") {
      writer.uint32(42).string(message.groupName);
    }
    if (!message.amountOut.isZero()) {
      writer.uint32(48).uint64(message.amountOut);
    }
    if (!message.amountIn.isZero()) {
      writer.uint32(56).uint64(message.amountIn);
    }
    if (!message.amountTotal.isZero()) {
      writer.uint32(120).uint64(message.amountTotal);
    }
    if (!message.revenueOut.isZero()) {
      writer.uint32(64).uint64(message.revenueOut);
    }
    if (!message.revenueIn.isZero()) {
      writer.uint32(72).uint64(message.revenueIn);
    }
    if (!message.revenueTotal.isZero()) {
      writer.uint32(112).uint64(message.revenueTotal);
    }
    if (!message.countOut.isZero()) {
      writer.uint32(80).uint64(message.countOut);
    }
    if (!message.countIn.isZero()) {
      writer.uint32(88).uint64(message.countIn);
    }
    if (!message.countTotal.isZero()) {
      writer.uint32(96).uint64(message.countTotal);
    }
    if (!message.capacity.isZero()) {
      writer.uint32(104).uint64(message.capacity);
    }
    if (message.turnover !== 0) {
      writer.uint32(133).float(message.turnover);
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): AggregatedForwards {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseAggregatedForwards } as AggregatedForwards;
    message.channels = [];
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.channels.push(ChanInfo.decode(reader, reader.uint32()));
          break;
        case 2:
          message.groupType = reader.int32() as any;
          break;
        case 4:
          message.groupId = reader.string();
          break;
        case 5:
          message.groupName = reader.string();
          break;
        case 6:
          message.amountOut = reader.uint64() as Long;
          break;
        case 7:
          message.amountIn = reader.uint64() as Long;
          break;
        case 15:
          message.amountTotal = reader.uint64() as Long;
          break;
        case 8:
          message.revenueOut = reader.uint64() as Long;
          break;
        case 9:
          message.revenueIn = reader.uint64() as Long;
          break;
        case 14:
          message.revenueTotal = reader.uint64() as Long;
          break;
        case 10:
          message.countOut = reader.uint64() as Long;
          break;
        case 11:
          message.countIn = reader.uint64() as Long;
          break;
        case 12:
          message.countTotal = reader.uint64() as Long;
          break;
        case 13:
          message.capacity = reader.uint64() as Long;
          break;
        case 16:
          message.turnover = reader.float();
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): AggregatedForwards {
    const message = { ...baseAggregatedForwards } as AggregatedForwards;
    message.channels = (object.channels ?? []).map((e: any) =>
      ChanInfo.fromJSON(e)
    );
    message.groupType =
      object.groupType !== undefined && object.groupType !== null
        ? groupTypeFromJSON(object.groupType)
        : 0;
    message.groupId =
      object.groupId !== undefined && object.groupId !== null
        ? String(object.groupId)
        : "";
    message.groupName =
      object.groupName !== undefined && object.groupName !== null
        ? String(object.groupName)
        : "";
    message.amountOut =
      object.amountOut !== undefined && object.amountOut !== null
        ? Long.fromString(object.amountOut)
        : Long.UZERO;
    message.amountIn =
      object.amountIn !== undefined && object.amountIn !== null
        ? Long.fromString(object.amountIn)
        : Long.UZERO;
    message.amountTotal =
      object.amountTotal !== undefined && object.amountTotal !== null
        ? Long.fromString(object.amountTotal)
        : Long.UZERO;
    message.revenueOut =
      object.revenueOut !== undefined && object.revenueOut !== null
        ? Long.fromString(object.revenueOut)
        : Long.UZERO;
    message.revenueIn =
      object.revenueIn !== undefined && object.revenueIn !== null
        ? Long.fromString(object.revenueIn)
        : Long.UZERO;
    message.revenueTotal =
      object.revenueTotal !== undefined && object.revenueTotal !== null
        ? Long.fromString(object.revenueTotal)
        : Long.UZERO;
    message.countOut =
      object.countOut !== undefined && object.countOut !== null
        ? Long.fromString(object.countOut)
        : Long.UZERO;
    message.countIn =
      object.countIn !== undefined && object.countIn !== null
        ? Long.fromString(object.countIn)
        : Long.UZERO;
    message.countTotal =
      object.countTotal !== undefined && object.countTotal !== null
        ? Long.fromString(object.countTotal)
        : Long.UZERO;
    message.capacity =
      object.capacity !== undefined && object.capacity !== null
        ? Long.fromString(object.capacity)
        : Long.UZERO;
    message.turnover =
      object.turnover !== undefined && object.turnover !== null
        ? Number(object.turnover)
        : 0;
    return message;
  },

  toJSON(message: AggregatedForwards): unknown {
    const obj: any = {};
    if (message.channels) {
      obj.channels = message.channels.map((e) =>
        e ? ChanInfo.toJSON(e) : undefined
      );
    } else {
      obj.channels = [];
    }
    message.groupType !== undefined &&
      (obj.groupType = groupTypeToJSON(message.groupType));
    message.groupId !== undefined && (obj.groupId = message.groupId);
    message.groupName !== undefined && (obj.groupName = message.groupName);
    message.amountOut !== undefined &&
      (obj.amountOut = (message.amountOut || Long.UZERO).toString());
    message.amountIn !== undefined &&
      (obj.amountIn = (message.amountIn || Long.UZERO).toString());
    message.amountTotal !== undefined &&
      (obj.amountTotal = (message.amountTotal || Long.UZERO).toString());
    message.revenueOut !== undefined &&
      (obj.revenueOut = (message.revenueOut || Long.UZERO).toString());
    message.revenueIn !== undefined &&
      (obj.revenueIn = (message.revenueIn || Long.UZERO).toString());
    message.revenueTotal !== undefined &&
      (obj.revenueTotal = (message.revenueTotal || Long.UZERO).toString());
    message.countOut !== undefined &&
      (obj.countOut = (message.countOut || Long.UZERO).toString());
    message.countIn !== undefined &&
      (obj.countIn = (message.countIn || Long.UZERO).toString());
    message.countTotal !== undefined &&
      (obj.countTotal = (message.countTotal || Long.UZERO).toString());
    message.capacity !== undefined &&
      (obj.capacity = (message.capacity || Long.UZERO).toString());
    message.turnover !== undefined && (obj.turnover = message.turnover);
    return obj;
  },

  fromPartial<I extends Exact<DeepPartial<AggregatedForwards>, I>>(
    object: I
  ): AggregatedForwards {
    const message = { ...baseAggregatedForwards } as AggregatedForwards;
    message.channels =
      object.channels?.map((e) => ChanInfo.fromPartial(e)) || [];
    message.groupType = object.groupType ?? 0;
    message.groupId = object.groupId ?? "";
    message.groupName = object.groupName ?? "";
    message.amountOut =
      object.amountOut !== undefined && object.amountOut !== null
        ? Long.fromValue(object.amountOut)
        : Long.UZERO;
    message.amountIn =
      object.amountIn !== undefined && object.amountIn !== null
        ? Long.fromValue(object.amountIn)
        : Long.UZERO;
    message.amountTotal =
      object.amountTotal !== undefined && object.amountTotal !== null
        ? Long.fromValue(object.amountTotal)
        : Long.UZERO;
    message.revenueOut =
      object.revenueOut !== undefined && object.revenueOut !== null
        ? Long.fromValue(object.revenueOut)
        : Long.UZERO;
    message.revenueIn =
      object.revenueIn !== undefined && object.revenueIn !== null
        ? Long.fromValue(object.revenueIn)
        : Long.UZERO;
    message.revenueTotal =
      object.revenueTotal !== undefined && object.revenueTotal !== null
        ? Long.fromValue(object.revenueTotal)
        : Long.UZERO;
    message.countOut =
      object.countOut !== undefined && object.countOut !== null
        ? Long.fromValue(object.countOut)
        : Long.UZERO;
    message.countIn =
      object.countIn !== undefined && object.countIn !== null
        ? Long.fromValue(object.countIn)
        : Long.UZERO;
    message.countTotal =
      object.countTotal !== undefined && object.countTotal !== null
        ? Long.fromValue(object.countTotal)
        : Long.UZERO;
    message.capacity =
      object.capacity !== undefined && object.capacity !== null
        ? Long.fromValue(object.capacity)
        : Long.UZERO;
    message.turnover = object.turnover ?? 0;
    return message;
  },
};

const baseChannelIDs: object = { chanIds: Long.UZERO };

export const ChannelIDs = {
  encode(
    message: ChannelIDs,
    writer: _m0.Writer = _m0.Writer.create()
  ): _m0.Writer {
    writer.uint32(10).fork();
    for (const v of message.chanIds) {
      writer.uint64(v);
    }
    writer.ldelim();
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): ChannelIDs {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseChannelIDs } as ChannelIDs;
    message.chanIds = [];
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          if ((tag & 7) === 2) {
            const end2 = reader.uint32() + reader.pos;
            while (reader.pos < end2) {
              message.chanIds.push(reader.uint64() as Long);
            }
          } else {
            message.chanIds.push(reader.uint64() as Long);
          }
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): ChannelIDs {
    const message = { ...baseChannelIDs } as ChannelIDs;
    message.chanIds = (object.chanIds ?? []).map((e: any) =>
      Long.fromString(e)
    );
    return message;
  },

  toJSON(message: ChannelIDs): unknown {
    const obj: any = {};
    if (message.chanIds) {
      obj.chanIds = message.chanIds.map((e) => (e || Long.UZERO).toString());
    } else {
      obj.chanIds = [];
    }
    return obj;
  },

  fromPartial<I extends Exact<DeepPartial<ChannelIDs>, I>>(
    object: I
  ): ChannelIDs {
    const message = { ...baseChannelIDs } as ChannelIDs;
    message.chanIds = object.chanIds?.map((e) => Long.fromValue(e)) || [];
    return message;
  },
};

const basePeerIDs: object = { pubKeys: "" };

export const PeerIDs = {
  encode(
    message: PeerIDs,
    writer: _m0.Writer = _m0.Writer.create()
  ): _m0.Writer {
    for (const v of message.pubKeys) {
      writer.uint32(10).string(v!);
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): PeerIDs {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...basePeerIDs } as PeerIDs;
    message.pubKeys = [];
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.pubKeys.push(reader.string());
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): PeerIDs {
    const message = { ...basePeerIDs } as PeerIDs;
    message.pubKeys = (object.pubKeys ?? []).map((e: any) => String(e));
    return message;
  },

  toJSON(message: PeerIDs): unknown {
    const obj: any = {};
    if (message.pubKeys) {
      obj.pubKeys = message.pubKeys.map((e) => e);
    } else {
      obj.pubKeys = [];
    }
    return obj;
  },

  fromPartial<I extends Exact<DeepPartial<PeerIDs>, I>>(object: I): PeerIDs {
    const message = { ...basePeerIDs } as PeerIDs;
    message.pubKeys = object.pubKeys?.map((e) => e) || [];
    return message;
  },
};

const baseTagIDs: object = { tagIds: "" };

export const TagIDs = {
  encode(
    message: TagIDs,
    writer: _m0.Writer = _m0.Writer.create()
  ): _m0.Writer {
    for (const v of message.tagIds) {
      writer.uint32(10).string(v!);
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): TagIDs {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseTagIDs } as TagIDs;
    message.tagIds = [];
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.tagIds.push(reader.string());
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): TagIDs {
    const message = { ...baseTagIDs } as TagIDs;
    message.tagIds = (object.tagIds ?? []).map((e: any) => String(e));
    return message;
  },

  toJSON(message: TagIDs): unknown {
    const obj: any = {};
    if (message.tagIds) {
      obj.tagIds = message.tagIds.map((e) => e);
    } else {
      obj.tagIds = [];
    }
    return obj;
  },

  fromPartial<I extends Exact<DeepPartial<TagIDs>, I>>(object: I): TagIDs {
    const message = { ...baseTagIDs } as TagIDs;
    message.tagIds = object.tagIds?.map((e) => e) || [];
    return message;
  },
};

const baseChanInfo: object = {
  chanId: Long.UZERO,
  alias: "",
  pubKey: "",
  closed: false,
  tags: "",
};

export const ChanInfo = {
  encode(
    message: ChanInfo,
    writer: _m0.Writer = _m0.Writer.create()
  ): _m0.Writer {
    if (!message.chanId.isZero()) {
      writer.uint32(8).uint64(message.chanId);
    }
    if (message.alias !== "") {
      writer.uint32(18).string(message.alias);
    }
    if (message.pubKey !== "") {
      writer.uint32(26).string(message.pubKey);
    }
    if (message.closed === true) {
      writer.uint32(40).bool(message.closed);
    }
    for (const v of message.tags) {
      writer.uint32(34).string(v!);
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): ChanInfo {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseChanInfo } as ChanInfo;
    message.tags = [];
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.chanId = reader.uint64() as Long;
          break;
        case 2:
          message.alias = reader.string();
          break;
        case 3:
          message.pubKey = reader.string();
          break;
        case 5:
          message.closed = reader.bool();
          break;
        case 4:
          message.tags.push(reader.string());
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): ChanInfo {
    const message = { ...baseChanInfo } as ChanInfo;
    message.chanId =
      object.chanId !== undefined && object.chanId !== null
        ? Long.fromString(object.chanId)
        : Long.UZERO;
    message.alias =
      object.alias !== undefined && object.alias !== null
        ? String(object.alias)
        : "";
    message.pubKey =
      object.pubKey !== undefined && object.pubKey !== null
        ? String(object.pubKey)
        : "";
    message.closed =
      object.closed !== undefined && object.closed !== null
        ? Boolean(object.closed)
        : false;
    message.tags = (object.tags ?? []).map((e: any) => String(e));
    return message;
  },

  toJSON(message: ChanInfo): unknown {
    const obj: any = {};
    message.chanId !== undefined &&
      (obj.chanId = (message.chanId || Long.UZERO).toString());
    message.alias !== undefined && (obj.alias = message.alias);
    message.pubKey !== undefined && (obj.pubKey = message.pubKey);
    message.closed !== undefined && (obj.closed = message.closed);
    if (message.tags) {
      obj.tags = message.tags.map((e) => e);
    } else {
      obj.tags = [];
    }
    return obj;
  },

  fromPartial<I extends Exact<DeepPartial<ChanInfo>, I>>(object: I): ChanInfo {
    const message = { ...baseChanInfo } as ChanInfo;
    message.chanId =
      object.chanId !== undefined && object.chanId !== null
        ? Long.fromValue(object.chanId)
        : Long.UZERO;
    message.alias = object.alias ?? "";
    message.pubKey = object.pubKey ?? "";
    message.closed = object.closed ?? false;
    message.tags = object.tags?.map((e) => e) || [];
    return message;
  },
};

export interface torqrpc {
  GetAggrigatedForwards(
    request: DeepPartial<AggregatedForwardsRequest>,
    metadata?: grpc.Metadata
  ): Promise<AggregatedForwardsResponse>;
}

export class torqrpcClientImpl implements torqrpc {
  private readonly rpc: Rpc;

  constructor(rpc: Rpc) {
    this.rpc = rpc;
    this.GetAggrigatedForwards = this.GetAggrigatedForwards.bind(this);
  }

  GetAggrigatedForwards(
    request: DeepPartial<AggregatedForwardsRequest>,
    metadata?: grpc.Metadata
  ): Promise<AggregatedForwardsResponse> {
    return this.rpc.unary(
      torqrpcGetAggrigatedForwardsDesc,
      AggregatedForwardsRequest.fromPartial(request),
      metadata
    );
  }
}

export const torqrpcDesc = {
  serviceName: "torqrpc.torqrpc",
};

export const torqrpcGetAggrigatedForwardsDesc: UnaryMethodDefinitionish = {
  methodName: "GetAggrigatedForwards",
  service: torqrpcDesc,
  requestStream: false,
  responseStream: false,
  requestType: {
    serializeBinary() {
      return AggregatedForwardsRequest.encode(this).finish();
    },
  } as any,
  responseType: {
    deserializeBinary(data: Uint8Array) {
      return {
        ...AggregatedForwardsResponse.decode(data),
        toObject() {
          return this;
        },
      };
    },
  } as any,
};

interface UnaryMethodDefinitionishR
  extends grpc.UnaryMethodDefinition<any, any> {
  requestStream: any;
  responseStream: any;
}

type UnaryMethodDefinitionish = UnaryMethodDefinitionishR;

interface Rpc {
  unary<T extends UnaryMethodDefinitionish>(
    methodDesc: T,
    request: any,
    metadata: grpc.Metadata | undefined
  ): Promise<any>;
}

export class GrpcWebImpl {
  private host: string;
  private options: {
    transport?: grpc.TransportFactory;

    debug?: boolean;
    metadata?: grpc.Metadata;
  };

  constructor(
    host: string,
    options: {
      transport?: grpc.TransportFactory;

      debug?: boolean;
      metadata?: grpc.Metadata;
    }
  ) {
    this.host = host;
    this.options = options;
  }

  unary<T extends UnaryMethodDefinitionish>(
    methodDesc: T,
    _request: any,
    metadata: grpc.Metadata | undefined
  ): Promise<any> {
    const request = { ..._request, ...methodDesc.requestType };
    const maybeCombinedMetadata =
      metadata && this.options.metadata
        ? new BrowserHeaders({
            ...this.options?.metadata.headersMap,
            ...metadata?.headersMap,
          })
        : metadata || this.options.metadata;
    return new Promise((resolve, reject) => {
      grpc.unary(methodDesc, {
        request,
        host: this.host,
        metadata: maybeCombinedMetadata,
        transport: this.options.transport,
        debug: this.options.debug,
        onEnd: function (response) {
          if (response.status === grpc.Code.OK) {
            resolve(response.message);
          } else {
            const err = new Error(response.statusMessage) as any;
            err.code = response.status;
            err.metadata = response.trailers;
            reject(err);
          }
        },
      });
    });
  }
}

type Builtin =
  | Date
  | Function
  | Uint8Array
  | string
  | number
  | boolean
  | undefined;

export type DeepPartial<T> = T extends Builtin
  ? T
  : T extends Long
  ? string | number | Long
  : T extends Array<infer U>
  ? Array<DeepPartial<U>>
  : T extends ReadonlyArray<infer U>
  ? ReadonlyArray<DeepPartial<U>>
  : T extends {}
  ? { [K in keyof T]?: DeepPartial<T[K]> }
  : Partial<T>;

type KeysOfUnion<T> = T extends T ? keyof T : never;
export type Exact<P, I extends P> = P extends Builtin
  ? P
  : P & { [K in keyof P]: Exact<P[K], I[K]> } & Record<
        Exclude<keyof I, KeysOfUnion<P>>,
        never
      >;

if (_m0.util.Long !== Long) {
  _m0.util.Long = Long as any;
  _m0.configure();
}
