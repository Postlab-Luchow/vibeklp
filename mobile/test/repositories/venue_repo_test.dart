import 'dart:convert';
import 'dart:io';

import 'package:dio/dio.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:mocktail/mocktail.dart';
import 'package:klp_guide/repositories/venue_repo.dart';

class MockDio extends Mock implements Dio {}

void main() {
  late MockDio dio;
  late VenueRepository repo;

  setUpAll(() {
    registerFallbackValue(RequestOptions(path: ''));
  });

  setUp(() {
    dio = MockDio();
    repo = VenueRepository(dio);
  });

  test('fetchAll returns parsed venues on 200', () async {
    const fixturePath = 'test/fixtures/venues_response_sample.json';
    final body =
        jsonDecode(File(fixturePath).readAsStringSync())
            as Map<String, dynamic>;

    when(() => dio.get<Map<String, dynamic>>('/api/venues')).thenAnswer(
      (_) async => Response<Map<String, dynamic>>(
        requestOptions: RequestOptions(path: '/api/venues'),
        statusCode: 200,
        data: body,
      ),
    );

    final venues = await repo.fetchAll();
    expect(venues, hasLength(1));
    expect(venues.first.id, 'venue-4044317c59e8f143');
    expect(venues.first.name, 'SCHLANZE');
  });

  test('fetchAll rethrows DioException on network failure', () async {
    when(() => dio.get<Map<String, dynamic>>('/api/venues')).thenThrow(
      DioException(
        requestOptions: RequestOptions(path: '/api/venues'),
        type: DioExceptionType.connectionTimeout,
      ),
    );

    await expectLater(repo.fetchAll(), throwsA(isA<DioException>()));
  });
}
