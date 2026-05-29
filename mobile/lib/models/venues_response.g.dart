// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'venues_response.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

_$VenuesResponseImpl _$$VenuesResponseImplFromJson(Map<String, dynamic> json) =>
    _$VenuesResponseImpl(
      venues: (json['venues'] as List<dynamic>)
          .map((e) => Venue.fromJson(e as Map<String, dynamic>))
          .toList(),
      total: (json['total'] as num).toInt(),
    );

Map<String, dynamic> _$$VenuesResponseImplToJson(
  _$VenuesResponseImpl instance,
) => <String, dynamic>{
  'venues': instance.venues.map((e) => e.toJson()).toList(),
  'total': instance.total,
};
