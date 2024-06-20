export interface WebsocketMessage {
  type?: string;
  data?: unknown;
}

export interface Event {
  event: string;
  data: unknown;
}

export interface Question {
  ask?: string;
}

export interface WorkOrder {
  order?: string;
  details?: unknown;
}

export interface JoinedRoomData {
  room_id?: number /* uint */;
  chat_log: UserMessageData[];
  name?: string;
  visitors: Visitor[];
}

export interface Visitor {
  id?: number /* uint */;
  name?: string;
}

export interface ErrorData {
  status_code?: number /* int */;
  message?: string;
  public?: boolean;
}

export interface UserLoggedInData {
  name?: string;
  id?: number /* uint */;
  access_token?: string;
}

export interface StreamIDUserNameData {
  stream_id?: string;
  name?: string;
}

export interface DirectMessageData {
  text?: string;
  to_user_id?: number /* uint */;
  from_user_id?: number /* uint */;
}

export interface UserMessageData {
  text?: string;
  from_user_name?: string;
  from_user_id?: number /* uint */;
  user_verified: boolean;
  to_user_id?: number /* uint */;
}

export interface UserMessageWorkOrderDetail {
  text?: string;
  to_user_id?: number /* uint */;
}

export interface UserMessageWorkOrder {
  order?: string;
  details?: UserMessageWorkOrderDetail;
}

export interface UserExitedChatData {
  name?: string;
  id?: number /* uint */;
}

export type CurrentGuestsData = CurrentGuest[];

export interface CurrentGuest {
  name?: string;
  id?: number /* uint */;
}

