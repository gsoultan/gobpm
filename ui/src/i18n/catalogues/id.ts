/**
 * Bahasa Indonesia.
 *
 * A second catalogue exists so the machinery is exercised by something real
 * rather than by a copy of English: a language with no plural inflection proves
 * that the plural forms come from `Intl.PluralRules` and not from an assumption
 * baked into the code. Indonesian has one form, so `other` carries every count.
 *
 * Incomplete by design — it covers exactly the keys that have been migrated.
 * A key with no entry here shows the key itself rather than silently falling
 * back to English, which is what makes the gap visible.
 */
import type { Catalogue } from '../translate';

const id: Catalogue = {
  'nav.dashboard': 'Dasbor',
  'nav.inbox': 'Kotak Masuk Saya',
  'nav.allTasks': 'Semua Tugas',
  'nav.processes': 'Proses',
  'nav.decisions': 'Keputusan',
  'nav.connectors': 'Konektor',
  'nav.instances': 'Instansi',
  'nav.projects': 'Proyek',
  'nav.organizations': 'Organisasi',
  'nav.groups': 'Grup',
  'nav.people': 'Orang',
  'nav.platformAccess': 'Akses platform',
  'nav.sectionWork': 'Pekerjaan',
  'nav.sectionBuild': 'Bangun',
  'nav.sectionOperate': 'Operasikan',
  'nav.sectionAdminister': 'Kelola',
  'nav.collapse': 'Ciutkan',
  'nav.mainLabel': 'Navigasi utama',

  'login.welcome': 'Selamat datang kembali',
  'login.subtitle': 'Masuk untuk melanjutkan ke ruang kerja Anda',
  'login.username': 'Nama pengguna',
  'login.password': 'Kata sandi',
  'login.submit': 'Masuk',
  'login.failedTitle': 'Tidak dapat masuk',
  'login.failedHelp':
    'Periksa nama pengguna dan kata sandi Anda, atau minta administrator menyetel ulang.',

  'inbox.title': 'Kotak Masuk Tugas',
  'inbox.subtitle': 'Kelola dan selesaikan tugas untuk {name}.',
  'inbox.assignedToMe': 'Ditugaskan ke Saya',
  'inbox.availableToClaim': 'Tersedia untuk Diambil',
  'inbox.search': 'Cari tugas…',
  'inbox.columnTask': 'Info Tugas',
  'inbox.columnAssignment': 'Penugasan',
  'inbox.columnAbout': 'Tentang apa',
  'inbox.columnTimeline': 'Lini masa',
  'inbox.columnStatus': 'Status',
  'inbox.columnActions': 'Tindakan',
  'inbox.claim': 'Ambil Tugas',
  'inbox.complete': 'Selesaikan',
  'inbox.noDueDate': 'Tanpa tenggat',
  'inbox.reference': 'Referensi {code}',
  'inbox.allCaughtUp': 'Semua sudah selesai',
  // One form for every count: Indonesian does not inflect for number.
  'inbox.nothingWaiting':
    '{count, plural, =0 {Tidak ada yang memerlukan perhatian Anda} other {# tugas menunggu}}',

  'offline.title': 'Anda sedang luring',
  'offline.body':
    'Anda masih dapat membaca yang sudah dimuat, dan menyelesaikan tugas di kotak masuk. Semuanya akan dikirim saat Anda kembali daring.',
  'offline.savedTitle': 'Disimpan di perangkat ini',
  'offline.queued': '{count, plural, other {# perubahan menunggu untuk dikirim}}',
  'offline.sending': 'Sedang dikirim. Disimpan di sini sampai server mengonfirmasi setiap satu.',
  'offline.keptHere': 'Disimpan di perangkat ini sampai Anda kembali daring.',
  'offline.tryAgain': 'Coba lagi sekarang',
  'update.title': 'Versi baru siap',
  'update.body':
    'Muat ulang saat Anda berada di titik berhenti yang tepat. Apa pun yang sedang Anda ketik belum tersimpan.',
  'update.reload': 'Muat ulang sekarang',
  'update.later': 'Nanti',

  'common.cancel': 'Batal',
  'common.save': 'Simpan',
  'common.delete': 'Hapus',
  'common.close': 'Tutup',
  'common.retry': 'Coba lagi',
  'common.loading': 'Memuat…',
  'common.language': 'Bahasa',
};

export default id;
