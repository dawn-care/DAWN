import { useState, useEffect } from 'react';
import AudioScreen from './audio';
import GreetingScreen from './greeting';
import TextViewerScreen from './text';
import PhotoViewerScreen from './photo';
import React from 'react';
import { Platform, Pressable, SafeAreaView, StyleSheet, Text, View } from 'react-native';
import * as SecureStore from 'expo-secure-store';
import { router } from 'expo-router';

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

export type MediaPlaybackProps = {
    onPlaybackComplete: () => void;
};

function GlassCard({ children, style }: { children: React.ReactNode; style?: any }) {
    return <View style={[styles.glassBase, style]}>{children}</View>;
}

function getHeaderTitle(item: ConfigurationWithMetadata): string {
    if (item.module_type === 'audio') {
        const metadata = item.metadata as AudioMetadata | null;
        return metadata?.displayed_name ?? 'Audio';
    }

    if (item.module_type === 'video') {
        const metadata = item.metadata as VideoMetadata | null;
        return metadata?.displayed_name ?? 'Video';
    }

    if (item.module_type === 'photo') {
        const metadata = item.metadata as PhotoMetadata | null;
        return metadata?.displayed_name ?? 'Photos';
    }

    if (item.module_type === 'message') {
        const metadata = item.metadata as MessageMetadata | null;
        return metadata?.message ?? 'Photos';
    }

    if (item.module_type === 'weather') {
        return 'Weather';
    }

    if (item.module_type === 'sports') {
        return 'Sports News';
    }

    if (item.module_type === 'news') {
        return 'Good News';
    }

    const metadata = item.metadata as EventMetadata | null;
    return metadata?.time ?? 'Event';
}

async function getStoredCarepartnerID(): Promise<number | null> {
    if (Platform.OS === 'web') {
        const id = localStorage.getItem('carepartnerID');
        return id ? parseInt(id, 10) : null;
    }
    const id = await SecureStore.getItemAsync('carepartnerID');
    return id ? parseInt(id, 10) : null;
}

export default function AgendaView(){
    const [items, setItems] = useState<ConfigurationWithMetadata[]>([]);
    const [currentIndex, setCurrentIndex] = useState(0);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);

    const fetchData = async () => {
        try{
            setLoading(true);

            // const carepartnerId = await getStoredCarepartnerID();
            // if (!carepartnerId) {
            //     setError('No carepartner ID found. Please pair this device first.');
            //     return;
            // }

            const { dayOfMonth, month, dayOfWeek } = currentDateToComponents();

            const params = new URLSearchParams({
                carepartner_id: "1",
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

        if (item.module_type === 'audio' || item.module_type === 'video') {
            return;
        }

        const timer = setTimeout(() => {
            setCurrentIndex((prev) => {
                if (prev + 1 >= items.length) {
                    router.replace('/');
                    return prev;
                }
                return prev + 1;
            });
        }, (item.duration || 10) * 1000);

        return () => clearTimeout(timer);
    }, [currentIndex, items]);

    const goToPreviousItem = () => {
        setCurrentIndex((prev) => Math.max(prev - 1, 0));
    };

    const goToNextItem = () => {
        setCurrentIndex((prev) => {
            if (prev + 1 >= items.length) {
                router.replace('/');
                return prev;
            }
            return prev + 1;
        });
    };

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

    const canGoPrev = currentIndex > 0;
    const canGoNext = items.length > 0;
    const headerTitle = getHeaderTitle(currentItem);

    return (
        <SafeAreaView style={styles.safe}>
            <View style={styles.topBar}>
                <Text style={styles.topTitle}>{headerTitle}</Text>
            </View>

            <View style={styles.container}>
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
                    <GreetingScreen
                        configuration={currentItem}
                        onPlaybackComplete={goToNextItem}
                    />
                )}

                {currentItem.module_type === "audio" && (
                    <AudioScreen
                        configuration={currentItem}
                        onPlaybackComplete={goToNextItem}
                    />
                )}

                <Pressable
                    onPress={goToPreviousItem}
                    disabled={!canGoPrev}
                    style={({ pressed }) => [
                        styles.leftWrapper,
                        pressed && canGoPrev && styles.arrowPressed,
                        !canGoPrev && styles.arrowDisabled,
                    ]}
                >
                    <GlassCard style={styles.arrowGlass}>
                        <Text style={styles.arrowText}>‹</Text>
                    </GlassCard>
                </Pressable>

                <Pressable
                    onPress={goToNextItem}
                    disabled={!canGoNext}
                    style={({ pressed }) => [
                        styles.rightWrapper,
                        pressed && canGoNext && styles.arrowPressed,
                        !canGoNext && styles.arrowDisabled,
                    ]}
                >
                    <GlassCard style={styles.arrowGlass}>
                        <Text style={styles.arrowText}>›</Text>
                    </GlassCard>
                </Pressable>
            </View>
        </SafeAreaView>
    );
}

const styles = StyleSheet.create({
    safe: {
        flex: 1,
        backgroundColor: '#000',
    },
    topBar: {
        backgroundColor: '#c7f3f6',
        borderBottomWidth: 2,
        borderBottomColor: 'rgba(0,0,0,0.6)',
        paddingHorizontal: 16,
        paddingVertical: 14,
        alignItems: 'center',
        justifyContent: 'center',
    },
    topTitle: {
        fontSize: 54,
        fontWeight: '300',
        color: 'rgba(0,0,0,0.85)',
        textAlign: 'center',
    },
    container: {
        flex: 1,
        position: 'relative',
    },
    glassBase: {
        borderRadius: 16,
        borderWidth: 1.5,
        borderColor: 'rgba(0,0,0,0.25)',
        backgroundColor: 'rgba(255,255,255,0.55)',
    },
    leftWrapper: {
        position: 'absolute',
        left: 26,
        top: '50%',
        marginTop: -41,
    },
    rightWrapper: {
        position: 'absolute',
        right: 26,
        top: '50%',
        marginTop: -41,
    },
    arrowGlass: {
        width: 82,
        height: 82,
        borderRadius: 18,
        borderWidth: 2,
        borderColor: 'rgba(0, 255, 240, 0.9)',
        alignItems: 'center',
        justifyContent: 'center',
    },
    arrowText: {
        fontSize: 54,
        color: 'rgba(0,0,0,0.75)',
        lineHeight: 54,
        marginTop: -2,
    },
    arrowPressed: {
        transform: [{ scale: 0.98 }],
    },
    arrowDisabled: {
        opacity: 0.45,
    },
});