import 'package:ffi/ffi.dart';
import 'package:flutter/material.dart';
import 'constants.dart';
import 'xgo.dart';

void main() {
  runApp(const MyApp());
}

class MyApp extends StatelessWidget {

  const MyApp({Key key}) : super(key:key);
  
  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'XGO mobile test',
      theme: ThemeData(
        primarySwatch: Colors.blue,
      ),
      home: const MyHomePage(title: 'XGO mobile test'),
    );
  }
}

class MyHomePage extends StatefulWidget {
  const MyHomePage({Key key, this.title}) : super(key: key);

  final String title;

  @override
  _MyHomePageState createState() => _MyHomePageState();
}

class _MyHomePageState extends State<MyHomePage> {
  final XGOBridge _xgoBridge = XGOBridge();

  void _show(message) {
    showDialog(
        builder: (ctx) => AlertDialog(content: Text(message.toString())),
        context: context);
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: Text(widget.title),
      ),
      body: Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: <Widget>[
            ElevatedButton(
                child: const Text('SingleComm'),
                onPressed: () async {
                  _xgoBridge.registerLogger(wrappedPrintPointer);
                  final _newBuffer =_xgoBridge.newBuffer(bufferSize);
                  final callRes = _xgoBridge.singleComm
                    (communicationPort, _newBuffer, bufferSize);
                  if (callRes != 1) {
                    _show('call has failed');
                  } else {
                    _show('call succeeded. Buffer is:\n'
                        '${_newBuffer.toDartString()}');
                  }
                  _xgoBridge.freeBuffer(_newBuffer);
                }),
          ],
        ),
      ),
    );
  }
}