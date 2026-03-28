import Link from "next/link";
import { 
  UserCheck, 
  GraduationCap, 
  Briefcase, 
  BookOpen 
} from "lucide-react";

export default function ApplyPage() {
  const membershipTypes = [
    {
      id: "asil",
      title: "Asil Üyelik",
      description: "Derneğin temel üyelik türü. Aktif katılım ve oylama hakkı sağlar.",
      icon: UserCheck,
      color: "bg-blue-500",
      hoverColor: "hover:bg-blue-600",
    },
    {
      id: "akademik",
      title: "Akademik Üyelik",
      description: "Akademik kariyer ve araştırma geçmişine sahip üyeler için.",
      icon: GraduationCap,
      color: "bg-purple-500",
      hoverColor: "hover:bg-purple-600",
    },
    {
      id: "profesyonel",
      title: "Profesyonel Üyelik",
      description: "Profesyonel iş deneyimine sahip üyeler için.",
      icon: Briefcase,
      color: "bg-green-500",
      hoverColor: "hover:bg-green-600",
    },
    {
      id: "ogrenci",
      title: "Öğrenci Üyelik",
      description: "Üniversite öğrencileri için özel üyelik türü.",
      icon: BookOpen,
      color: "bg-orange-500",
      hoverColor: "hover:bg-orange-600",
    },
  ];

  return (
    <div className="min-h-screen bg-gradient-to-br from-slate-50 to-slate-100">
      <div className="container mx-auto px-4 py-16">
        {/* Header */}
        <div className="text-center mb-12">
          <h1 className="text-4xl font-bold text-slate-900 mb-4">
            Üyelik Başvurusu
          </h1>
          <p className="text-lg text-slate-600 max-w-2xl mx-auto">
            Derneğimize katılmak için uygun üyelik türünü seçin ve başvuru
            sürecinizi başlatın.
          </p>
        </div>

        {/* Membership Type Cards */}
        <div className="grid grid-cols-1 md:grid-cols-2 gap-6 max-w-5xl mx-auto">
          {membershipTypes.map((type) => (
            <Link
              key={type.id}
              href={`/apply/${type.id}`}
              className="group"
            >
              <div className="bg-white rounded-lg shadow-md hover:shadow-xl transition-all duration-300 p-8 border-2 border-transparent hover:border-slate-200">
                <div className="flex items-start space-x-4">
                  <div
                    className={`${type.color} ${type.hoverColor} p-3 rounded-lg text-white transition-colors`}
                  >
                    <type.icon className="w-8 h-8" />
                  </div>
                  <div className="flex-1">
                    <h2 className="text-2xl font-semibold text-slate-900 mb-2 group-hover:text-slate-700 transition-colors">
                      {type.title}
                    </h2>
                    <p className="text-slate-600">{type.description}</p>
                    <div className="mt-4 text-sm font-medium text-slate-900 group-hover:text-slate-700 flex items-center">
                      Başvuru Yap
                      <svg
                        className="w-4 h-4 ml-1 group-hover:translate-x-1 transition-transform"
                        fill="none"
                        stroke="currentColor"
                        viewBox="0 0 24 24"
                      >
                        <path
                          strokeLinecap="round"
                          strokeLinejoin="round"
                          strokeWidth={2}
                          d="M9 5l7 7-7 7"
                        />
                      </svg>
                    </div>
                  </div>
                </div>
              </div>
            </Link>
          ))}
        </div>

        {/* Info Section */}
        <div className="mt-16 max-w-3xl mx-auto bg-white rounded-lg shadow-md p-8">
          <h3 className="text-xl font-semibold text-slate-900 mb-4">
            Başvuru Süreci Hakkında
          </h3>
          <ul className="space-y-3 text-slate-600">
            <li className="flex items-start">
              <span className="text-blue-500 font-bold mr-2">•</span>
              <span>
                Başvurunuz sistem tarafından incelenir ve değerlendirilir.
              </span>
            </li>
            <li className="flex items-start">
              <span className="text-blue-500 font-bold mr-2">•</span>
              <span>
                Asil ve Akademik üyelik başvuruları için en az 3 referans
                gereklidir.
              </span>
            </li>
            <li className="flex items-start">
              <span className="text-blue-500 font-bold mr-2">•</span>
              <span>
                Profesyonel ve Öğrenci üyelik başvuruları danışma sürecinden
                geçer.
              </span>
            </li>
            <li className="flex items-start">
              <span className="text-blue-500 font-bold mr-2">•</span>
              <span>
                Başvuru durumunuz hakkında e-posta yoluyla bilgilendirileceksiniz.
              </span>
            </li>
          </ul>
        </div>
      </div>
    </div>
  );
}
