import 'package:flutter/material.dart';

import 'features/home/home_screen.dart';

class KlpGuideApp extends StatelessWidget {
  const KlpGuideApp({super.key});

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'KLP-Guide',
      debugShowCheckedModeBanner: false,
      theme: ThemeData(useMaterial3: true, colorSchemeSeed: Colors.deepOrange),
      home: const HomeScreen(),
    );
  }
}
