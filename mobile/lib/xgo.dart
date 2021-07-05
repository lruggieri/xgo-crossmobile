import 'dart:ffi';
import 'dart:io';
import 'package:ffi/ffi.dart';

typedef SingleComFuncDart =
    int Function(int port, Pointer<Utf8> buffer, int bufferLen);
typedef NewBuffer = Pointer<Utf8> Function(int bufferLen);
typedef FreeBuffer = void Function(Pointer<Utf8> buffer);
typedef RegisterLogger = void Function(Pointer loggerFunction);

void wrappedPrint(Pointer<Utf8> arg){
  print(arg.toDartString());
}
typedef WrappedPrintC = Void Function(Pointer<Utf8> a);
final wrappedPrintPointer = Pointer.fromFunction<WrappedPrintC>(wrappedPrint);

class XGOBridge {
  SingleComFuncDart _singleComm;
  NewBuffer _newBuffer;
  FreeBuffer _freeBuffer;
  RegisterLogger _registerLogger;

  XGOBridge() {
    final xgoLib = Platform.isAndroid
        ? DynamicLibrary.open('libmain-android-22-arm64.so')
        : DynamicLibrary.process();

    _singleComm =
    xgoLib
        .lookup<NativeFunction<Int8 Function(Int32, Pointer<Utf8>, Int32)
    >>('ServerClientSingleCommunication')
        .asFunction();

    _newBuffer =
      xgoLib
          .lookup<NativeFunction<Pointer<Utf8> Function(Int32 bufferLen)
      >>('NewStringBuffer')
          .asFunction();

    _freeBuffer =
        xgoLib
            .lookup<NativeFunction<Void Function(Pointer<Utf8> buffer)
        >>('FreeBuffer')
            .asFunction();

    _registerLogger =
        xgoLib
            .lookup<NativeFunction<Void Function(Pointer loggerFunction)
        >>('RegisterLogger')
            .asFunction();

  }

  int singleComm(int port, Pointer<Utf8> buffer, int bufferLen) =>
      _singleComm(port, buffer, bufferLen);
  Pointer<Utf8> newBuffer(int bufferLen) => _newBuffer(bufferLen);
  void freeBuffer(Pointer<Utf8> buffer) => _freeBuffer(buffer);
  void registerLogger(Pointer loggerFunction) =>
      _registerLogger(loggerFunction);
}
