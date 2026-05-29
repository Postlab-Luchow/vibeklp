// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'venue.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

_$VenueImpl _$$VenueImplFromJson(Map<String, dynamic> json) => _$VenueImpl(
  id: json['id'] as String,
  name: json['name'] as String,
  description: json['description'] as String? ?? '',
  address: Address.fromJson(json['address'] as Map<String, dynamic>),
  coordinates: Coordinates.fromJson(
    json['coordinates'] as Map<String, dynamic>,
  ),
  contact: json['contact'] == null
      ? const Contact()
      : Contact.fromJson(json['contact'] as Map<String, dynamic>),
  amenities:
      (json['amenities'] as List<dynamic>?)?.map((e) => e as String).toList() ??
      const <String>[],
  categories:
      (json['categories'] as List<dynamic>?)
          ?.map((e) => VenueCategory.fromJson(e as Map<String, dynamic>))
          .toList() ??
      const <VenueCategory>[],
  bikeRoute: json['bikeRoute'] as String? ?? '',
  source: json['source'] as String? ?? '',
  eventIds:
      (json['eventIds'] as List<dynamic>?)?.map((e) => e as String).toList() ??
      const <String>[],
  exhibitionIds:
      (json['exhibitionIds'] as List<dynamic>?)
          ?.map((e) => e as String)
          .toList() ??
      const <String>[],
  eventCount: (json['eventCount'] as num?)?.toInt() ?? 0,
  exhibitionCount: (json['exhibitionCount'] as num?)?.toInt() ?? 0,
);

Map<String, dynamic> _$$VenueImplToJson(_$VenueImpl instance) =>
    <String, dynamic>{
      'id': instance.id,
      'name': instance.name,
      'description': instance.description,
      'address': instance.address.toJson(),
      'coordinates': instance.coordinates.toJson(),
      'contact': instance.contact.toJson(),
      'amenities': instance.amenities,
      'categories': instance.categories.map((e) => e.toJson()).toList(),
      'bikeRoute': instance.bikeRoute,
      'source': instance.source,
      'eventIds': instance.eventIds,
      'exhibitionIds': instance.exhibitionIds,
      'eventCount': instance.eventCount,
      'exhibitionCount': instance.exhibitionCount,
    };

_$AddressImpl _$$AddressImplFromJson(Map<String, dynamic> json) =>
    _$AddressImpl(
      street: json['street'] as String? ?? '',
      postalCode: json['postalCode'] as String? ?? '',
      city: json['city'] as String? ?? '',
    );

Map<String, dynamic> _$$AddressImplToJson(_$AddressImpl instance) =>
    <String, dynamic>{
      'street': instance.street,
      'postalCode': instance.postalCode,
      'city': instance.city,
    };

_$CoordinatesImpl _$$CoordinatesImplFromJson(Map<String, dynamic> json) =>
    _$CoordinatesImpl(
      lat: (json['lat'] as num).toDouble(),
      lng: (json['lng'] as num).toDouble(),
    );

Map<String, dynamic> _$$CoordinatesImplToJson(_$CoordinatesImpl instance) =>
    <String, dynamic>{'lat': instance.lat, 'lng': instance.lng};

_$ContactImpl _$$ContactImplFromJson(Map<String, dynamic> json) =>
    _$ContactImpl(
      phone: json['phone'] as String? ?? '',
      email: json['email'] as String? ?? '',
      website: json['website'] as String? ?? '',
    );

Map<String, dynamic> _$$ContactImplToJson(_$ContactImpl instance) =>
    <String, dynamic>{
      'phone': instance.phone,
      'email': instance.email,
      'website': instance.website,
    };

_$VenueCategoryImpl _$$VenueCategoryImplFromJson(Map<String, dynamic> json) =>
    _$VenueCategoryImpl(
      name: json['name'] as String,
      dates:
          (json['dates'] as List<dynamic>?)?.map((e) => e as String).toList() ??
          const <String>[],
    );

Map<String, dynamic> _$$VenueCategoryImplToJson(_$VenueCategoryImpl instance) =>
    <String, dynamic>{'name': instance.name, 'dates': instance.dates};
