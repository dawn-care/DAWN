import { useState, useEffect } from 'react';
import AudioScreen from './audio';
import GreetingScreen from './greeting';
import TextViewerScreen from './text';
import PhotoViewerScreen from './photo';
import React from 'react';
import { View } from 'react-native';

// Metadata shapes returned inline by the backend
export type PhotoMetadata = { id: number; uri: string; displayed_name: string; filename: string; extension: string };
export type VideoMetadata = { id: number; uri: string; displayed_name: string; filename: string; extension: string };
export type AudioMetadata = { id: number; uri: string; displayed_name: string; filename: string; extension: string };
export type WeatherMetadata = { ID: number; ZIPCODE: string };
export type NewsMetadata = { ID: number; CATEGORY: string; LANGUAGE: string; COUNTRY: string };
export type MessageMetadata = { id: number; message: string; image_uri: string };
export type EventMetadata = { id: number; list_id: number; time: string; text: string };
export type CarepartnerLocationMetadata = { id: number; location: string; message: string; image_uri: string };
export type CaregiverLocationMetadata = { id: number; location: string; message: string; image_uri: string };

// Matches the ConfigurationWithMetadata Go struct returned by GetConfigurationsByDayAndCarepartner
export type ConfigurationWithMetadata = {
    id: number;
    caregiver_id: number;
    carepartner_id: number;
    module_id: number;
    occurence_type: string;
    occurence_definition: string[];
    position: number;
    duration: number;
    module_type: string;
    metadata_id: number;
    metadata: PhotoMetadata | VideoMetadata | AudioMetadata | WeatherMetadata
        | NewsMetadata | MessageMetadata | EventMetadata
        | CarepartnerLocationMetadata | CaregiverLocationMetadata | null;
};

export default function AgendaView(){
    const [items, setItems] = useState<ConfigurationWithMetadata[]>([]);
    const [currentIndex, setCurrentIndex] = useState(0);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);

    const fetchData = async () => {
        try{
            setLoading(true);

            const { dayOfMonth, month, dayOfWeek } = currentDateToComponents();

            const params = new URLSearchParams({
                carepartner_id: '1',
                day: String(dayOfMonth),
                month: String(month),
                day_of_week: String(dayOfWeek),
            })

            const res = await fetch(`http://localhost:8080/configurations/day?${params}`);
            if (!res.ok) throw new Error('Failed to fetch configurations');
            const data: ConfigurationWithMetadata[] = await res.json();
            console.log('Fetched configurations:', data);
            setItems(data || []);
        } catch (err) {
            const errorMessage = err instanceof Error ? err.message : 'An error occurred';
            setError(errorMessage);
            console.error('Error fetching data:', err);
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        fetchData();
    }, [])
    
    useEffect(() => {
        if (!items.length) return;

        const item = items[currentIndex];
        console.log('Displaying item:', item);

        const timer = setTimeout(() => {
            setCurrentIndex((prev) => (prev + 1) % items.length);
        }, (item.duration || 10) * 1000);

        return () => clearTimeout(timer);
    }, [currentIndex, items]);

    const currentDateToComponents = () => {
        const currentDate = new Date();
        const dayOfWeek = currentDate.getDay();
        const dayOfMonth = currentDate.getDate();
        const month = currentDate.getMonth() + 1;
        return { dayOfMonth, month, dayOfWeek };
    }

    const currentItem = items[currentIndex];

    if (loading) {
        return <View />;
    }

    if (!currentItem) {
        return <View />;
    }

    return (
        <View>
            {(currentItem.module_type === "photo" || currentItem.module_type === "message") && (
                <PhotoViewerScreen configuration={currentItem} />
            )}

            {(currentItem.module_type === "weather" ||
            currentItem.module_type === "sports" ||
            currentItem.module_type === "news" ||
            currentItem.module_type === "event") && (
                <TextViewerScreen configuration={currentItem} />
            )}

            {currentItem.module_type === "video" && (
                <GreetingScreen configuration={currentItem} />
            )}

            {currentItem.module_type === "audio" && (
                <AudioScreen configuration={currentItem} />
            )}
        </View>
    );
}