var terminalType = 'pc'
var currentAppKey = 'seeCloud'
var screenLineCount = 500
var remoteTimeout = 300
var timerCount = 0
var changeFlag = false
var hasEnerRoom = false
var userInfo = {
  userName: '',
  uId: -1,
}
var wsAddr = 'ws://127.0.0.1:8688/debug/chatroom'

/************初始化******************/
$(function () {
  $('#remoteMessage').height($(window).height() - 130)
  initColor('backgroundColor', '#F9F9F9')
  initColor('color', '#000080')
  $('#close').attr('disabled', true)
  var roomId = ''
  // $('#room_id').val(roomId)
  // $('#room_id').bind('input propertychange', function (event) {
  //   changeFlag = true
  // })
})

function addLine() {
  addMessage('<hr>')
}

function addMessage(message) {
  var content = $('#order').dialog('options').content
  content = content == null ? '' : content
  $('#order').dialog({ content: content + message })
}
// 打开命令选择框
function openOrder() {
  //获取系统命令参数
  //getTerminalType();
  // 初始化选择命令
  initDefaultCmd()
  intGetParam()
  $('#order').dialog('open')
}

//区别每个窗口会话
var sessionId = new Date().valueOf()
//是否连接
var isConnect = false

//灰化连接按钮
function disable(disable) {
  if (disable) {
    $('#user_name').attr('disabled', true)
    $('#connect').attr('disabled', true)
    $('#close').attr('disabled', false)
  } else {
    $('#mac').attr('disabled', false)
    $('#connect').attr('disabled', false)
    $('#close').attr('disabled', true)
  }
}
//关闭连接
function closeSession() {
  EndText()
  if (socket) {
    socket.close()
  }
  isConnected = false

  isConnect = false
  disable(false)
  rawMessage('<br>Close success!<br>')
}

var msgFunc = null
var paramNo = 0

//屏幕打印消息
function screenMessage(platform, message, ret) {
  if (platform == 'client' || platform == 'local') {
    platform = '[' + platform + '] #'
  } else {
    platform = '[' + platform + ']#'
  }
  message = message.replace(/(^\s*)|(\s*$)/g, '').replace(/(\n)/g, '<br>')
  if (
    message != undefined &&
    message != null &&
    message != '' &&
    message != '<br>'
  ) {
    $('#remoteMessage').append(
      '<div class="order-line' +
        (ret != 0 ? ' error' : '') +
        '"><span style="padding-right:5px;float: left;">' +
        platform +
        '</span><div>' +
        message +
        '</div></div>'
    )
    $('#remoteMessage').scrollTop($('#remoteMessage')[0].scrollHeight)
  }
}
function rawMessage(message) {
  if (!message) {
    return
  }
  message = message.toString()
  message = message.replace(/(^\s*)|(\s*$)/g, '').replace(/(\n)/g, '<br>')
  if (
    message != undefined &&
    message != null &&
    message != '' &&
    message != '<br>'
  ) {
    $('#remoteMessage').append('<div class="order-line">' + message + '</div>')
    $('#remoteMessage').scrollTop($('#remoteMessage')[0].scrollHeight)
  }
  //保证只有3000行，防止页面溢出
  if ($('#remoteMessage').children('.order-line').length > screenLineCount) {
    $('#remoteMessage .order-line').first().remove()
  }
}

//背景色
function changeColor(input) {
  $('#remoteMessage').css(input.id, input.value)
  //$.cookie(input.id + '_${userName}', input.value);
}

function initColor(id, defaultColor) {
  //var color = $.cookie(id + '_${userName}');
  //color = color ? color : defaultColor;
  var color = defaultColor
  $('#remoteMessage').css(id, color)
  $('#' + id).val(color)
}
//清屏
function clearMessage() {
  $('#remoteMessage').html('')
}

/************ websocket 通信 ******************/
var boardIp = ''
var openId = ''
var socket
var ClientAction = {}
ClientAction.Error = -1 // 通用错误，发生错误时，Data请传递`ClientError`结构体
ClientAction.Message = 0 // 普通消息
ClientAction.CreateRoom = 1 // 创建房间
ClientAction.JoinRoom = 2 // 加入房间
ClientAction.ChangedRoom = 3 // 房间信息变更
ClientAction.GetRoomInfo = 4 // 获取房间信息
ClientAction.StartFrameSync = 5 // 开启帧同步
ClientAction.StopFrameSync = 6 // 停止帧同步
ClientAction.UploadFrame = 7 // 上传帧同步数据
ClientAction.Login = 8 // 登陆(预留用作将来鉴权)
ClientAction.FData = 9 // 帧数据
ClientAction.RoomMessage = 10 // 发送房间消息
ClientAction.GetRoomList = 11 // 获取房间列表
ClientAction.ExitRoom = 12 //退出房间

var hasConnected = false
function reConnect() {
  if (hasConnected) {
    socket.close()
    hasConnected = false
  }
  checkConnect()
}

// 连接终端
function checkConnect() {
  console.log('Enter CheckConnect')
  if (!$.trim($('#user_name').val())) {
    screenMessage('local', '用户名不可为空!', -1)
    return
  }
  try {
    socket = new WebSocket(wsAddr)
  } catch (e) {
    alert('error')
    return
  }
  socket.onopen = sOpen
  socket.onerror = sError
  socket.onmessage = sMessage
  socket.onclose = sClose
  hasConnected = true

  disable(true)
}

function sOpen() {
  rawMessage('connect success!')
  registAndLogin()
}
function sError(e) {
  console.log('On Error ')
  console.log(e)
  alert('On Error ' + e)
}

var reader = {
  readAs: function (type, blob, cb) {
    var r = new FileReader()
    r.onloadend = function () {
      if (typeof cb === 'function') {
        cb.call(r, r.result)
      }
    }
    try {
      r['readAs' + type](blob)
    } catch (e) {}
  },
}

var fileData = []
function parseBlob(blob) {
  if (blob.size < 1024) {
    fileData.push(blob)

    var fileName = 'ynbs_log.txt'
    saveLogToFolder(fileName)
  } else {
    fileData.push(blob)
  }
}

function sMessage(msg) {
  //console.log(msg)
  if (typeof msg.data != 'string') {
    parseBlob(msg.data)
    return
  }
  jsonMsg = JSON.parse(msg.data)
  if (jsonMsg.Data) {
    if (jsonMsg.Op == ClientAction.CreateRoom) {
      getRoomList()
    }
    if (jsonMsg.Op == ClientAction.JoinRoom) {
      console.log(jsonMsg.Data)
      if (jsonMsg.Data.Code == 0) {
        hasEnerRoom = true
      }
      rawMessage(jsonMsg.Data.Msg)
      getRoomList()
    } else if (jsonMsg.Op == ClientAction.ExitRoom) {
      if (jsonMsg.Data.Code == 0) {
        hasEnerRoom = false
        console.log(jsonMsg.Data.Msg)
        rawMessage(jsonMsg.Data.Msg)
      } else {
        rawMessage(jsonMsg.Data)
      }
      getRoomList()
    } else if (jsonMsg.Op == ClientAction.GetRoomList) {
      //console.log(jsonMsg)
      if (jsonMsg.Data.Code == 0) {
        $('#room_list').html('')
        for (var idx in jsonMsg.Data.Data) {
          $('<li>' + jsonMsg.Data.Data[idx] + '</li>')
            .appendTo('#room_list')
            .bind('click', function () {
              if (confirm('是否加入房间' + jsonMsg.Data.Data[idx])) {
                var strRoomId = jsonMsg.Data.Data[idx].toString()
                joinRoomById(strRoomId)
              }
            })
        }
      }
    } else if (jsonMsg.Op == ClientAction.RoomMessage) {
      console.log(jsonMsg)
      // finalData =
      //   JSON.stringify(jsonMsg.Data.Sender) +
      //   ':' +
      //   JSON.stringify(jsonMsg.Data.Msg)
      finalData = '[' + jsonMsg.Data.Sender + '] : ' + jsonMsg.Data.Msg
      console.log(finalData)
      rawMessage(finalData)
    } else if (jsonMsg.Op == ClientAction.Message) {
      rawMessage(jsonMsg.Data)
    } else {
      rawMessage(JSON.stringify(jsonMsg.Data))
    }
  }
}
function sClose(e) {
  rawMessage('connect closed:' + e.code)
  screenMessage('local', '连接断开!', -1)
}

function Logout() {
  socket.close()
}

function sendMessage() {
  socket.send(JSON.stringify($('#message').val()))
  $('#message').val('')
}

function sendRoomMessage() {
  var msgData = {
    Op: ClientAction.RoomMessage,
    Data: {
      Sender: userInfo.userName,
      Msg: $('#message').val(),
    },
  }
  socket.send(JSON.stringify(msgData))
  rawMessage('[' + userInfo.userName + '] : ' + $('#message').val())
  $('#message').val('')
}

function Close() {
  socket.close()
}
function validateIP(str) {
  const re =
    /^(1\d{2}|2[0-4]\d|25[0-5]|[1-9]\d|[0-9])\.((1\d{2}|2[0-4]\d|25[0-5]|[1-9]\d|\d)\.){2}(1\d{2}|2[0-4]\d|25[0-5]|[1-9]\d|\d)$/
  return re.test(str)
}

function registAndLogin() {
  // if (!$.trim($('#room_id').val())) {
  //   screenMessage('local', 'ROOM ID不可为空!', -1)
  //   return
  // }
  userInfo.userName = $('#user_name').val()
  var msgData = {
    Op: ClientAction.Login,
    Data: { UserName: userInfo.userName, OpenID: '' },
  }
  socket.send(JSON.stringify(msgData))
  $('#message').val()
}

function createRoom() {
  var msgData = {
    Op: ClientAction.CreateRoom,
  }
  console.log(JSON.stringify(msgData))
  socket.send(JSON.stringify(msgData))
}

function getRoomList() {
  var msgData = {
    Op: ClientAction.GetRoomList,
  }
  socket.send(JSON.stringify(msgData))
}

function joinRoom() {
  roomId = $('#room_id').val()
  var msgData = {
    Op: ClientAction.JoinRoom,
    Data: { RoomId: roomId },
  }
  socket.send(JSON.stringify(msgData))
}

function joinRoomById(roomId) {
  var msgData = {
    Op: ClientAction.JoinRoom,
    Data: { RoomId: roomId },
  }
  socket.send(JSON.stringify(msgData))
}

function exitRoom() {
  var msgData = {
    Op: ClientAction.ExitRoom,
  }
  socket.send(JSON.stringify(msgData))
  hasEnerRoom = false
}

window.onbeforeunload = function (e) {
  console.log('退出浏览器清理')
  alert('退出')
  if (socket) {
    socket.close()
  }
  isConnected = false
  hasEnerRoom = false
  alert('退出')
}
