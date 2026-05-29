import 'package:flutter/material.dart';

class KlpGuideApp extends StatelessWidget {
  const KlpGuideApp({super.key});

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'KLP-Guide',
      debugShowCheckedModeBanner: false,
      theme: ThemeData(useMaterial3: true, colorSchemeSeed: Colors.deepOrange),
      home: const Scaffold(
        body: Center(child: Text('KLP-Guide bootstraps.')),
      ),
    );
  }
}
